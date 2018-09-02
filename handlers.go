package main

import "net/http"

func receiveTransfer(w http.ResponseWriter, r *http.Request) {
	data, err := parse(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("got invalid data at /transfer")
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
}

func receiveTransferAck(w http.ResponseWriter, r *http.Request) {
	data, err := parse(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("got invalid data at /transfer/ack")
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
}
