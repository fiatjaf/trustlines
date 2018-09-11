package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

type User struct {
	Id string `db:"id"`
}

type Signable interface {
	signerId() string
	signature() []byte
	verifiablePayload() []byte
	signablePayload() []byte
}

type Timestamps struct {
	StHere  time.Time `db:"st_here" json:"st_here"`
	StThere time.Time `db:"st_there" json:"st_there"`
}

func (t Timestamps) get(w string) time.Time {
	switch w {
	case "here":
		return t.StHere
	case "there":
		return t.StThere
	}
	return time.Now()
}

type Transfer struct {
	Id       string `db:"id" json:"id"`
	Debtor   string `db:"debtor" json:"debtor"`
	Creditor string `db:"creditor" json:"creditor"`
	Timestamps
	ActualDate  time.Time `db:"actual_date" json:"actual_date"`
	Amount      int       `db:"amount" json:"amount"`
	Currency    string    `db:"currency" json:"currency"`
	Description string    `db:"description" json:"description"`
	Next        Next      `json:"-"`
	NextEnc     string    `db:"next" json:"next"`
	Signature   string    `db:"signature" json:"signature"`
}

func (t Transfer) checkValidity() {

}

func (t Transfer) signerId() string  { return t.Debtor }
func (t Transfer) signature() []byte { return []byte(t.Signature) }

func (t Transfer) payload(timestamp string) []byte {
	payload := t.Creditor + ":" + t.Debtor + ":" +
		strconv.FormatInt(
			t.Timestamps.get(timestamp).Unix(),
			10,
		) + ":" +
		strconv.Itoa(t.Amount) + ":" + t.Currency

	if !t.Next.End {
		payload += ":" + t.Next.OutgoingCreditor + ":" +
			strconv.Itoa(t.Next.AmountToForward) + ":" +
			t.Next.CurrencyToForward + ":" +
			t.Next.NextOnion
	}

	return []byte(payload)
}
func (t Transfer) verifiablePayload() []byte { return t.payload("there") }
func (t Transfer) signablePayload() []byte   { return t.payload("here") }

type Next struct {
	OutgoingCreditor  string `json:"outgoing_creditor,omitempty"`
	AmountToForward   int    `json:"amount_to_forward,omitempty"`
	CurrencyToForward string `json:"currency_to_forward,omitempty"`

	End bool `json:"end"` // true if there's no other info here

	NextOnion string `json:"next_onion,omitempty"` // encrypted to outgoing_creditor
}

func (n *Next) decode(enc string) error {
	ciphertext, err := hex.DecodeString(enc)
	if err != nil {
		return err
	}

	label := []byte("orders")

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader,
		privateKey, ciphertext, label)
	if err != nil {
		return err
	}

	err = json.Unmarshal(plaintext, n)
	if err != nil {
		return err
	}

	return nil
}

func (n *Next) encode(key *rsa.PublicKey) (string, error) {
	plaintext, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	enc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, plaintext, []byte{})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(enc), nil
}

type Ack struct {
	Creditor   string `db:"creditor" json:"creditor"`
	TransferId string `db:"transfer_id" json:"transfer_id"`
	Signature  string `db:"signature" json:"signature"`
	Timestamps
}

func (ack Ack) signerId() string  { return ack.Creditor }
func (ack Ack) signature() []byte { return []byte(ack.Signature) }

func (ack Ack) payload(timestamp string) []byte {
	payload := ack.Creditor + ":" + ack.TransferId + ":" +
		strconv.FormatInt(
			ack.Timestamps.get(timestamp).Unix(),
			10,
		)
	return []byte(payload)
}
func (ack Ack) verifiablePayload() []byte { return ack.payload("there") }
func (ack Ack) signablePayload() []byte   { return ack.payload("here") }

type Trustline struct {
	Truster  string `db:"truster" json:"truster"`
	Trusted  string `db:"trusted" json:"trusted"`
	Amount   int    `db:"amount" json:"amount"`
	Currency string `db:"currency" json:"amount"`
}
