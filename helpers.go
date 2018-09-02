package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)

var client = http.Client{Timeout: time.Second * 6}
var emptyResult = gjson.Result{}

func belongsHere(actor string) bool {
	u, err := url.Parse(actor)
	if err != nil {
		return false
	}

	return u.Host == s.Hostname
}

func parse(data io.Reader) (gjson.Result, error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return emptyResult, err
	}

	if gjson.ValidBytes(b) == false {
		return emptyResult, errors.New("got invalid content: " + string(b))
	}

	return gjson.ParseBytes(b), nil
}
