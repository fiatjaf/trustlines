package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)

func writeInbox(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	data := gjson.ParseBytes(b)

	actorId := data.Get("actor.id").String()
	if err := checkSignature(r, actorId); err != nil {
		http.Error(w, "invalid http signature", 403)
		return
	}

	typ := data.Get("type").String()
	obj := data.Get("object")
	objId := obj.Get("id").String()

	if len(obj.Array()) == 1 {
		// we only have the object id, let's fetch it
		obj, err = get(objId)
		if err != nil {
			log.Warn().Str("id", objId).Msg("failed to fetch object from origin")
			http.Error(w, "failed to fetch object from origin", 400)
			return
		}
	}

	objType := obj.Get("type").String()
	objCurrency := obj.Get("currency").String()
	objAmount := obj.Get("amount").Int()
	objDescription := obj.Get("description").String()
	objTimestamp := time.Unix(obj.Get("timestamp").Int(), 0)
	objSignature := obj.Get("signature").String()

	// these fields will have varying names depending on the type of object
	objSource := obj.Get("creditor").String()
	if objSource == "" {
		objSource = obj.Get("payer").String()
	}
	objTarget := obj.Get("debtor").String()
	if objTarget == "" {
		objTarget = obj.Get("payee").String()
	}

	// ensure a little consistency on the timestamps
	now := time.Now()
	if objTimestamp.After(now.Add(time.Second*30)) ||
		objTimestamp.Before(now.Add(time.Second*30*-1)) {
		log.Warn().Str("id", objId).Str("t", objTimestamp.Format(time.RFC3339)).
			Msg("timestamp received wrong")
		http.Error(w, "timestamp received wrong", 400)
		return
	}

	// optional fields (depending on the type of object)
	objActualDate := obj.Get("actualDate").String()
	objOverAmount := obj.Get("overAmount").String()
	objTimeout := obj.Get("timeout").Int()
	objCancelled := obj.Get("cancelled").Bool()
	objHash := obj.Get("hash").String()
	objNext := obj.Get("next").String()

	switch typ {
	case "Create":
		// someone from somewhere is declaring a debt to someone on this instance,
		// we'll just accept that.
		switch objType {
		case "https://trustlin.es/ns#Debt":
			if !belongsHere(objTarget) {
				err = errors.New("target doesn't belong here")
				goto end
			}

			_, err = pg.Exec(`
INSERT INTO debts
(id, source, target, currency, amount, description, t, signature, actual_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            `, objId, objSource, objTarget, objCurrency, objAmount, objDescription,
				objTimestamp, objSignature, objActualDate)
		case "https://trustlin.es/ns#Settlement":
			_, err = pg.Exec(`
INSERT INTO settlements
(id, source, target, currency, amount, description, t, signature, actual_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            `, objId, objSource, objTarget, objCurrency, objAmount, objDescription,
				objTimestamp, objSignature, objActualDate)
		case "https://trustlin.es/ns#Interest":
			_, err = pg.Exec(`
INSERT INTO interest_charges
(id, source, target, currency, amount, description, t, signature, actual_date, over_amount)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            `, objId, objSource, objTarget, objCurrency, objAmount, objDescription,
				objTimestamp, objSignature, objActualDate, objOverAmount)
		}

	case "Update":
		switch objType {
		case "https://trustlin.es/ns#Trustline":
		}

	case "Announce":
		// someone is offering us a hashed contract.
		// suppose it's an A->B->C payment
		// if we're A we'll never get this, we'll be sending the Announce instead
		// if we're B we create a contract to the actor specified in the "next" field
		//   (C) and send him an "Announce" and wait for a preimage with which we'll
		//   reply to A.
		// if we're C we can just reply with the preimage.
		switch objType {
		case "https://trustlin.es/ns#HashedContract":
			// TODO verify signature

			var preimage string
			if objNext == "" {
				// search preimage
				data, err := rds.HGetAll(objHash).Result()
				if err == nil {
					if data["currency"] == objCurrency && data["amount"] == objAmount {
						preimage = data["preimage"]
					} else {
						// this contract is not what we were expecting
						return
					}
				} else if objNext == "" {
					// reply with an error
					return
				}
			}

			// create new hashed contract, now targetting next
			_, err = pg.Exec(`
INSERT INTO hashed_contracts
(id, source, target, currency, amount, description, t, signature,
 timeout, cancelled, hash, preimage)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            `)

			_, err = pg.Exec(`
INSERT INTO hashed_contracts
(id, source, target, currency, amount, description, t, signature,
 timeout, cancelled, hash, preimage)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            `, objId, objSource, objTarget, objCurrency, objAmount, objDescription,
				objTimestamp, objSignature,
				objTimeout, objCancelled, objHash,
				preimage)

			if err != nil {
				log.Warn().Str("id", objId).
					Msg("failed to save hashed contract to database")
				http.Error(w, "failed to save hashed contract to database", 500)
				return
			}

			objUri := "https://" + s.Hostname + "/object/" + objId

			if preimage == "" {
				// get preimage from next and then reply with it
				post(objNext, map[string]interface{}{
					"@context": "https://www.w3.org/ns/activitystreams",
					"type":     "Announce",
					"object":   objId,
				})
			}

			w.Header.Set("Location")

			if preimage != "" {
				// reply with the preimage
			} else {
				// it's an error? we should wait for the hashed contract timeout, right?
			}
		}

	case "Offer":
		// someone is trying to get an user to acknowledge a debt,
		// we prompt the user. if he accepts  we send a Create with the debt object
		switch objType {
		case "https://trustlin.es/ns#Debt":
		case "https://trustlin.es/ns#Interest":
		case "https://trustlin.es/ns#Settlement":
		}
	}

end:
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Warn().Err(err).Msg("")
		return
	}
}

func readInbox(w http.ResponseWriter, r *http.Request) {
}

func writeOutbox(w http.ResponseWriter, r *http.Request) {
}

func readOutbox(w http.ResponseWriter, r *http.Request) {
}
