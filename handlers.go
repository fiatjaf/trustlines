package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func receiveTransfer(w http.ResponseWriter, r *http.Request) {
	var transfer Transfer
	err := json.NewDecoder(r.Body).Decode(&transfer)
	if err != nil {
		log.Warn().Err(err).Msg("got invalid data at /transfer")
		w.WriteHeader(400)
		return
	}

	log := log.With().
		Str("db", transfer.Debtor).
		Str("cr", transfer.Creditor).
		Int("amt", transfer.Amount).
		Str("curr", transfer.Currency).
		Bool("hop", !transfer.Next.End).
		Logger()

	// what for them means here, for us means there
	transfer.StThere = transfer.StHere

	// parse onion stuff
	err = transfer.Next.decode(transfer.NextEnc)
	if err != nil {
		log.Warn().Err(err).Msg("failed to decrypt and decode onion")
		w.WriteHeader(400)
		return
	}

	if err := checkTimestamps(transfer.Timestamps); err != nil {
		log.Warn().Err(err).Msg("transfer is too old")
		w.WriteHeader(408)
		return
	}

	// check if this server has the receiver account
	if !belongsHere(transfer.Creditor) {
		log.Warn().Err(err).Msg("got a payment for an account that is not here")
		w.WriteHeader(404)
		return
	}

	// check transfer signature
	if err := checkSignature(transfer); err != nil {
		// TODO blacklist this server?
		log.Warn().Err(err).Msg("got a payment with an invalid signature")
		w.WriteHeader(403)
		return
	}

	// check if this is a multihop payment and proceed accordingly
	if transfer.Next.End {
		// we're the final hop.
		// just accept this debt.
	} else {
		// we must create a new debt to transfer.Next.OutgoingCreditor

		// check if the currency to forward is the same as received
		// TODO: implement currency exchange
		if transfer.Next.CurrencyToForward != transfer.Currency {
			log.Info().Str("to_forward", transfer.Next.CurrencyToForward).
				Msg("got a payment with different currency to forward")
			w.WriteHeader(406)
			return
		}

		// check if we're receiving at least the same as we're forwarding
		// TODO: implement configurable fees
		if transfer.Amount < transfer.Next.AmountToForward {
			log.Info().Int("to_forward", transfer.Next.AmountToForward).
				Msg("got a payment with a greater amount to forward")
			w.WriteHeader(406)
			return
		}

		// check if we have trust on the sender
		avin, err := availableTrust(transfer.Debtor, transfer.Creditor, transfer.Currency)
		if err != nil {
			log.Warn().Err(err).Msg("failed to check available trust to here")
			w.WriteHeader(500)
			return
		} else if avin < transfer.Amount {
			log.Info().Int("available", avin).Msg("rejected as our trustlines are full")
			w.WriteHeader(406)
			return
		}

		// check if we have credit with the outgoing creditor
		avout, err := availableTrust(transfer.Creditor,
			transfer.Next.OutgoingCreditor, transfer.Next.CurrencyToForward)
		if err != nil {
			log.Warn().Err(err).Msg("failed to check available trust from here")
			w.WriteHeader(500)
			return
		} else if avout < transfer.Next.AmountToForward {
			log.Info().Int("available", avout).Msg("rejected as next trustlines are full")
			w.WriteHeader(406)
			return
		}

		// actually send the debt payment to the next in line
		go forwardPayment(transfer.Creditor, transfer.Description, transfer.Next)
	}

	// archive the debt payment we've received here
	// if there is a .Next in the line, then it is conditional
	// and we must present an ACK from the next peer to the peer
	// that is sending us this
	_, err = pg.Exec(`
INSERT INTO transfers
(id, debtor, creditor, description, amount, currency,
 st_there, actual_date, next,
 signature)
VALUES
($1, $2, $3, $4, $5, $6,
 $7, $8, $9, $10)
    `, transfer.Id, transfer.Debtor, transfer.Creditor,
		transfer.Description, transfer.Amount, transfer.Currency,
		transfer.StThere, transfer.ActualDate, transfer.NextEnc,
		transfer.Signature)

	w.WriteHeader(200)
}

func receiveTransferAck(w http.ResponseWriter, r *http.Request) {
	var ack Ack
	err := json.NewDecoder(r.Body).Decode(&ack)
	if err != nil {
		log.Warn().Err(err).Msg("got invalid data at /transfer/ack")
		w.WriteHeader(400)
		return
	}

	log := log.With().
		Str("cr", ack.Creditor).
		Str("trf", ack.TransferId).
		Logger()

	// what for them means here, for us means there
	ack.StThere = ack.StHere

	if err := checkTimestamps(ack.Timestamps); err != nil {
		log.Warn().Err(err).Msg("ack is too old")
		w.WriteHeader(408)
		return
	}

	// check if this server is the originator of the transfer
	if !belongsHere(ack.TransferId) {
		log.Warn().Err(err).Msg("got an ack for a transfer that is not here")
		w.WriteHeader(404)
		return
	}

	// check ack signature
	if err := checkSignature(ack); err != nil {
		// TODO blacklist this server?
		log.Warn().Err(err).Msg("got a payment with an invalid signature")
		w.WriteHeader(403)
		return
	}

	_, err = pg.Exec(`
INSERT INTO acks
(creditor, transfer_id, st_there, signature)
VALUES ($1, $2, $3, $4)
    `, ack.Creditor, ack.TransferId, ack.StHere, ack.Signature)

	w.WriteHeader(200)
}

func servePublicKey(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, s.PublicKey)
}
