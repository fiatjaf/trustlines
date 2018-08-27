package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx/types"
)

func serveUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var user User
	err := pg.Get(&user, "SELECT id FROM users WHERE id = $1", id)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "resource not found", 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	userIri := "https://" + s.Hostname + "/~/" + user.Id

	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":          "https://www.w3.org/ns/activitystreams",
		"id":                userIri,
		"type":              "Person",
		"inbox":             userIri + "/inbox",
		"outbox":            userIri + "/outbox",
		"preferredUsername": user.Id,
		"publicKey": map[string]interface{}{
			"id":           "https://" + s.Hostname + "/public-key",
			"owner":        userIri,
			"publicKeyPem": s.PublicKey,
		},
	})
}

func serveKey(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case "application/activity+json":
		json.NewEncoder(w).Encode(map[string]string{
			"@context":     "https://www.w3.org/ns/activitystreams",
			"id":           "https://" + s.Hostname + "/public-key",
			"type":         "Key",
			"publicKeyPem": s.PublicKey,
		})
	default:
		fmt.Fprint(w, s.PublicKey)
	}
}

func serveKind(kind string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		var data types.JSONText
		err := pg.Get(&data, "SELECT data FROM "+kind+" WHERE id = $1", id)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "resource not found", 404)
			} else {
				http.Error(w, err.Error(), 500)
			}
			return
		}

		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Error().Err(err).Str("data", data.String()).Msg("error encoding")
		}
	}
}
