package main

import (
	"net/url"
)

func belongsHere(actor string) bool {
	u, err := url.Parse(actor)
	if err != nil {
		return false
	}

	return u.Host == s.Hostname
}
