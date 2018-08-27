package main

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"

	"github.com/go-fed/httpsig"
)

func checkSignature(r *http.Request, actorId string) error {
	vef, err := httpsig.NewVerifier(r)
	if err != nil {
		return err
	}

	data, err := get(vef.KeyId())
	if err != nil {
		return err
	}

	ks := data.Get("publicKey.publicKeyPem").String()
	if ks == "" {
		ks = data.Get("publicKeyPem").String()
	}

	pk, err := decodePEM([]byte(ks))
	if err != nil {
		return err
	}

	err = vef.Verify(pk, httpsig.RSA_SHA256)
	if err != nil {
		return err
	}

	// check if this key is the correct one for this actor
	actor, err := get(actorId)
	if err != nil {
		return err
	}

	if actor.Get("publicKey.id").String() != vef.KeyId() {
		return errors.New("public key id doesn't match actor's public key")
	}

	return nil
}

func decodePEM(data []byte) (crypto.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("response is not a public key PEM block")
	}

	return x509.ParsePKIXPublicKey(block.Bytes)
}
