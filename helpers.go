package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)

var client = http.Client{Timeout: time.Second * 6}
var emptyResult = gjson.Result{}

func belongsHere(o string) bool {
	return extractServer(o) == s.Hostname
}

func extractServer(o string) string {
	u, _ := url.Parse(o)
	return u.Host
}

func checkTimestamps(t Timestamps) error {
	if t.StThere.Add(time.Second * 15).Before(time.Now()) {
		return errors.New("too old")
	}

	return nil
}

func availableTrust(from, to string, currency string) (int, error) {
	var total int
	err := pg.Get(&total, `
SELECT amount
FROM trustlines
WHERE truster = $1 AND trusted = $2 AND currency = $3
    `, from, to, currency)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	var used int
	err = pg.Get(&used, `
SELECT sum(v)
FROM (
  SELECT amount AS v
  FROM transfers
  WHERE currency = $3 AND debtor = $1 AND creditor = $2
UNION ALL
  SELECT -amount AS v
  FROM transfers
  WHERE currency = $3 AND debtor = $2 AND creditor = $1
) AS m
    `)
	if err != nil {
		return 0, err
	}

	return total - used, nil
}
