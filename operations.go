package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lucsky/cuid"
)

func sendDebtPayment() {

}

func forwardPayment(debtor string, desc string, n Next) {
	nextTransfer := Transfer{
		Id:          "https://" + s.Hostname + "/transfer/" + cuid.New(),
		Debtor:      debtor,
		Creditor:    n.OutgoingCreditor,
		Amount:      n.AmountToForward,
		Currency:    n.CurrencyToForward,
		Description: desc,
		NextEnc:     n.NextOnion,
	}
	nextTransfer.Timestamps.StHere = time.Now()

	logger := log.With().
		Str("d", debtor).Str("c", nextTransfer.Creditor).
		Int("amt", nextTransfer.Amount).Str("curr", nextTransfer.Currency).
		Str("id", nextTransfer.Id).
		Logger()

	if signature, err := sign(nextTransfer); err == nil {
		nextTransfer.Signature = signature
	} else {
		logger.Error().Err(err).Msg("failed to sign transfer on forwardPayment")
		return
	}

	targetServer := extractServer(n.OutgoingCreditor)
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(nextTransfer); err != nil {
		logger.Error().Err(err).Msg("failed to encode transfer on forwardPayment")
		return
	}

	_, err := http.Post("https://"+targetServer+"/pay", "application/json", body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to send transfer on forwardPayment")
		return
	}
}
