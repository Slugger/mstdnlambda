package jwt

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	jwtlib "github.com/dgrijalva/jwt-go/v4"
)

var ErrInvalidJwtSigningMethod = errors.New("unexpected signing method")
var ErrJwtParseFailure = errors.New("jwt parse failed")

func Verify(vjwt *VerifiableJwt) error {
	if _, err := jwtlib.Parse(vjwt.Token, vjwt.publicKey, jwtlib.WithAudience(vjwt.Aud)); err != nil {
		return fmt.Errorf("%w: %s", ErrJwtParseFailure, err.Error())
	}
	return nil
}

type VerifiableJwt struct {
	Token     string
	PublicKey *ecdsa.PublicKey
	Aud       string
}

func (vjwt *VerifiableJwt) publicKey(token *jwtlib.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwtlib.SigningMethodECDSA); !ok || token.Header["alg"] != "ES256" {
		return nil, ErrInvalidJwtSigningMethod
	}
	return vjwt.PublicKey, nil
}
