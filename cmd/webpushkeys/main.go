package main

/*
	The webpushkeys command is a simple command line utility that will generate
	a unique set of keys and secrets needed to successfully subscribe an endpoint
	to Mastodon push notifications.

	Running webpushkeys will generate a random set of keys and secerts and print
	them out as a json object to the console.  You MUST protect the private key
	and the shared key. Anyone with these secrets can decrypt the push
	notifications received from Mastodon. The shared key MUST only be shared
	with Mastodon when making the API call to subscribe to push notifications.
	The private key MUST NEVER be shared with anyone.  If either is compromised
	then you should immediately delete the subscription with the Mastodon instance
	and create a new one with new keys.
*/

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
)

type webPushKeys struct {
	PublicKey    string `json:"publicKey"`
	PrivateKey   string `json:"privateKey"`
	SharedSecret string `json:"sharedSecret"`
}

func main() {
	keys := genWebPushKeys()
	output, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", string(output))
}

func genWebPushKeys() *webPushKeys {
	private := genKey()
	privateBytes := private.D.Bytes()
	publicBytes := elliptic.Marshal(elliptic.P256(), private.PublicKey.X, private.PublicKey.Y)

	return &webPushKeys{
		PublicKey:    b64encode(publicBytes),
		PrivateKey:   b64encode(privateBytes),
		SharedSecret: b64encode(genSharedSecret()),
	}
}

func genSharedSecret() []byte {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func b64encode(input []byte) string {
	return b64.RawURLEncoding.EncodeToString(input)
}

func genKey() *ecdsa.PrivateKey {
	buf := make([]byte, 128)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(buf))
	if err != nil {
		panic(err)
	}
	return key
}
