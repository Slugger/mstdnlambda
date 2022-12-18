package jwt_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/slugger/mstdnlambda/internal/jwt"
	"github.com/stretchr/testify/assert"
)

func TestVerifyJwtSucceeds(t *testing.T) {
	privKey := genKey()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodES256, jwtlib.MapClaims{"aud": "https://foo.com", "exp": time.Now().Unix() + 600})
	encodedToken, err := token.SignedString(privKey)
	if err != nil {
		panic(err)
	}
	vtoken := jwt.VerifiableJwt{
		Token:     encodedToken,
		PublicKey: &privKey.PublicKey,
		Aud:       "https://foo.com",
	}
	err = jwt.Verify(&vtoken)
	assert.Nil(t, err)
}

func TestVerifyJwtFailsIfValidationFails(t *testing.T) {
	privKey := genKey()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodES256, jwtlib.MapClaims{"aud": "https://foo.com", "exp": 0})
	encodedToken, err := token.SignedString(privKey)
	if err != nil {
		panic(err)
	}
	vtoken := jwt.VerifiableJwt{
		Token:     encodedToken,
		PublicKey: &privKey.PublicKey,
		Aud:       "https://foo.com",
	}
	err = jwt.Verify(&vtoken)
	assert.ErrorIs(t, err, jwt.ErrJwtParseFailure)
}

func TestVerifyFailsIfTokenIsNotUsingES256SigningMethod(t *testing.T) {
	token := jwtlib.New(jwtlib.SigningMethodHS256) // method HS256 is not acceptable, must be ES256
	encodedToken, err := token.SignedString([]byte{})
	if err != nil {
		panic(err)
	}
	vtoken := jwt.VerifiableJwt{
		Token:     encodedToken,
		PublicKey: nil,
		Aud:       "foo",
	}
	err = jwt.Verify(&vtoken)
	assert.ErrorIs(t, err, jwt.ErrJwtParseFailure)
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
