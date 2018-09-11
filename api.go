package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lucsky/cuid"
)

func register(w http.ResponseWriter, r *http.Request) {
	user, err := d.VerifyAuth(r.Header.Get("Authorization"))
	if err != nil {
		log.Warn().Err(err).Msg("invalid accountd token")
		w.WriteHeader(401)
		return
	}

	_, err = pg.Exec(`INSERT INTO users (id) VALUES ($1)`, user)
	if err != nil {
		log.Warn().Err(err).Msg("failed to register user")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
}

func handleCreateDebt(w http.ResponseWriter, r *http.Request) {
	var transfer Transfer
	err := json.NewDecoder(r.Body).Decode(&transfer)
	if err != nil {
		log.Warn().Err(err).Msg("got invalid data at /create-debt")
		w.WriteHeader(400)
		return
	}

	if !belongsHere(transfer.Debtor) {
		log.Warn().Str("d", transfer.Debtor).Msg("debtor doesn't belong here")
		w.WriteHeader(403)
		return
	}

	if user, err := d.VerifyAuth(r.Header.Get("Authorization")); err != nil || user != transfer.Debtor {
		log.Warn().Str("d", transfer.Debtor).Err(err).Msg("debtor not authorized")
		w.WriteHeader(401)
		return
	}

	transfer.Id = "https://" + s.Hostname + "/~/" + cuid.Slug()
	transfer.Timestamps.StHere = time.Now()
	transfer.Next.End = true

	targetkey, err := fetchPublicKey(extractServer(transfer.Creditor))
	if err != nil {
		log.Error().Err(err).Str("c", transfer.Creditor).
			Msg("failed to fetch public key")
		w.WriteHeader(503)
		return
	}

	if enc, err := transfer.Next.encode(targetkey); err != nil {
		log.Error().Err(err).Msg("failed to encrypt onion")
		w.WriteHeader(500)
		return
	} else {
		transfer.NextEnc = enc
	}

	if signature, err := sign(transfer); err != nil {
		log.Error().Err(err).Msg("failed to sign")
		w.WriteHeader(500)
		return
	} else {
		transfer.Signature = signature
	}

	w.WriteHeader(202)
	return
}

func handleSendPayment(w http.ResponseWriter, r *http.Request) {

}
