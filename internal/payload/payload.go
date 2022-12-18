package payload

import (
	"fmt"

	ece "github.com/crow-misia/http-ece"
)

type EncryptedPayload struct {
	SharedSecret   []byte
	MyPrivateKey   []byte
	TheirPublicKey []byte
	Salt           []byte
	Data           []byte
}

func Decrypt(payload *EncryptedPayload) (string, error) {
	cleartxt, err := ece.Decrypt(payload.Data,
		ece.WithPrivate(payload.MyPrivateKey),
		ece.WithAuthSecret(payload.SharedSecret),
		ece.WithDh(payload.TheirPublicKey),
		ece.WithEncoding(ece.AESGCM),
		ece.WithSalt(payload.Salt))
	if err != nil {
		return "", fmt.Errorf("%s: %w", "payload decrypt failed", err)
	}
	return string(cleartxt), nil
}
