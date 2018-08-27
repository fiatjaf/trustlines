package main

import (
	"errors"

	"github.com/sheenobu/go-webfinger"
)

type wfresolver struct{}

func (_ wfresolver) FindUser(username string, hostname string, r []webfinger.Rel) (*webfinger.Resource, error) {
	if hostname != s.Hostname {
		return nil, errors.New("not found")
	}

	var user User
	err := pg.Get(&user, "SELECT id FROM users WHERE id = $1", username)
	if err != nil {
		return nil, err
	}

	return &webfinger.Resource{
		Subject: "acct:" + username + "@" + hostname,
		Links: []webfinger.Link{
			{
				Rel:  "self",
				Type: "application/activity+json",
				HRef: "https://" + s.Hostname + "/user/" + username,
			},
		},
	}, nil
}

func (_ wfresolver) DummyUser(username string, hostname string, r []webfinger.Rel) (*webfinger.Resource, error) {
	return nil, errors.New("not found")
}

func (_ wfresolver) IsNotFoundError(err error) bool { return true }
