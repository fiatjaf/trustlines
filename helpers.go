package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var client = http.Client{Timeout: time.Second * 6}
var emptyResult = gjson.Result{}

func parseUser(u string) (res struct {
	user   string
	server string
}) {
	parts := strings.Split(u, "@")
	if len(parts) != 2 {
		return
	}
	res.user = parts[0]
	res.server = parts[1]
	return
}

func extractServer(o string) string { return parseUser(o).server }
func extractUser(o string) string   { return parseUser(o).user }

func belongsHere(o string) bool {
	if extractServer(o) != s.Hostname {
		return false
	}

	var u string
	err := pg.Get(&u, `SELECT id FROM users WHERE id = $1`, extractUser(o))
	return err == nil
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
