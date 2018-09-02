package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/tidwall/gjson"
)

func getBytes(u string) ([]byte, error) {
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func get(u string) (gjson.Result, error) {
	resp, err := client.Get(u)
	if err != nil {
		return emptyResult, err
	}
	return parse(resp.Body)
}

func post(u string, data interface{}) (gjson.Result, error) {
	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(data)
	if err != nil {
		return emptyResult, err
	}

	resp, err := client.Post(u, "application/json", buffer)
	if err != nil {
		return emptyResult, err
	}
	return parse(resp.Body)
}
