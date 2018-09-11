package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
)

func checkSignature(s Signable) error {
	key, err := fetchPublicKey(extractServer(s.signerId()))
	if err != nil {
		return err
	}

	hash := sha256.Sum256(s.verifiablePayload())

	return rsa.VerifyPKCS1v15(
		key,
		crypto.SHA256,
		hash[:],
		s.signature(),
	)
}

func sign(s Signable) (string, error) {
	hash := sha256.Sum256(s.signablePayload())

	signature, err := rsa.SignPKCS1v15(
		nil,
		privateKey,
		crypto.SHA256,
		hash[:],
	)
	if err != nil {
		return "", err
	}

	return string(signature), nil
}

func fetchPublicKey(server string) (*rsa.PublicKey, error) {
	resp, err := http.Get("https://" + server + "/public-key")
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return decodePublicKeyPEM(b)
}

func decodePublicKeyPEM(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("not a public key PEM block")
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func decodePrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("not a private key PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
