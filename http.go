package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)

var client = http.Client{Timeout: time.Second * 6}
var emptyResult = gjson.Result{}

func get(u string) (gjson.Result, error) {
	urlp, _ := url.Parse(u)
	header := http.Header{}
	header.Set("Accept", "application/activity+json")

	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    urlp,
		Header: header,
	})

	if err != nil {
		return emptyResult, err
	}

	return handle(resp)
}

func handle(resp *http.Response) (gjson.Result, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emptyResult, err
	}

	if gjson.ValidBytes(b) == false {
		return emptyResult, errors.New("invalid response: " + string(b))
	}

	return gjson.ParseBytes(b), nil
}
