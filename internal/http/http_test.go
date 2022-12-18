package http_test

/*
	mstdnlambda
	Copyright (C) 2022 Battams, Derek <derek@battams.ca>

	This program is free software; you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation; either version 2 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License along
	with this program; if not, write to the Free Software Foundation, Inc.,
	51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/slugger/mstdnlambda/internal/cfg"
	"github.com/slugger/mstdnlambda/internal/http"
	"github.com/stretchr/testify/assert"
)

func TestExtractJwtSucceeds(t *testing.T) {
	expectedDomainName := "foo.com"
	pubkey := genP256PublicKey()
	pubkeyBytes := elliptic.Marshal(elliptic.P256(), pubkey.X, pubkey.Y)
	req := events.LambdaFunctionURLRequest{
		Headers: map[string]string{"authorization": "WebPush MyToken", "crypto-key": b64UrlEncodeKeyValBytes("p256ecdsa", pubkeyBytes)},
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: expectedDomainName,
		},
	}
	result, err := http.ExtractJwt(req)
	assert.Nil(t, err)
	assert.Equal(t, "MyToken", result.Token)
	assert.Equal(t, pubkeyBytes, elliptic.Marshal(elliptic.P256(), result.PublicKey.X, result.PublicKey.Y))
	assert.Equal(t, fmt.Sprintf("https://%s", expectedDomainName), result.Aud)
}

func TestExtractJwtFailsWhenPublicKeyIsNotBase64Encoded(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		Headers: map[string]string{"authorization": "WebPush MyToken", "crypto-key": "p256ecdsa=(##)"},
	}
	_, err := http.ExtractJwt(req)
	assert.ErrorIs(t, err, http.ErrNotBase64Encoded)
}

func TestExtractJwtFailsWhenPublicKeyIsInvalid(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		Headers: map[string]string{"authorization": "WebPush MyToken", "crypto-key": "foobar"},
	}
	_, err := http.ExtractJwt(req)
	assert.ErrorIs(t, err, http.ErrInvalidHeader)
	assert.ErrorContains(t, err, "invalid key/val header: [foobar]")
}

func TestExtractJwtFailsWhenPublicKeyIsMissing(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		Headers: map[string]string{"authorization": "WebPush MyToken"},
	}
	_, err := http.ExtractJwt(req)
	assert.ErrorIs(t, err, http.ErrMissingHeader)
	assert.ErrorContains(t, err, "crypto-key")
}

func TestExtractJwtFailsWhenBearerTokenIsMissing(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		Headers: map[string]string{},
	}
	_, err := http.ExtractJwt(req)
	assert.ErrorIs(t, err, http.ErrMissingHeader)
	assert.ErrorContains(t, err, "authorization")
}

func TestExtractJwtFailsWhenBearerTokenIsInvalid(t *testing.T) {
	testCases := []struct {
		hdrVal string
		desc   string
	}{
		{"Bearer MyToken", "invalid auth scheme"},
		{"webPush MyToken", "mismatched auth scheme"}, // scheme MUST be WebPush (case sensitive)
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req := events.LambdaFunctionURLRequest{
				Headers: map[string]string{"authorization": tc.hdrVal},
			}
			_, err := http.ExtractJwt(req)
			assert.ErrorIs(t, err, http.ErrInvalidHeader)
		})
	}
}

func TestExtractPayloadSucceeds(t *testing.T) {
	expectedSalt := "ValidSalt"
	expectedCryptoKey := "ValidCryptoKey"
	expectedBody := "ValidBody"
	expectedPrivateKey := "ValidPrivateKey"
	expectedSharedSecret := "ValidSharedSecret"

	initEnv(envValues{
		privateKey:   b64UrlEncode(expectedPrivateKey),
		sharedSecret: b64UrlEncode(expectedSharedSecret),
	})
	defer clearEnv()
	cfg.ParseConfig()

	req := events.LambdaFunctionURLRequest{
		IsBase64Encoded: true,
		Headers:         map[string]string{"encryption": b64UrlEncodeKeyVal("salt", expectedSalt), "crypto-key": b64UrlEncodeKeyVal("dh", expectedCryptoKey)},
		Body:            b64StdEncode(expectedBody),
	}

	result, err := http.ExtractPayload(req)
	assert.Nil(t, err)
	assert.Equal(t, expectedBody, string(result.Data))
	assert.Equal(t, expectedSalt, string(result.Salt))
	assert.Equal(t, expectedCryptoKey, string(result.TheirPublicKey))
	assert.Equal(t, expectedPrivateKey, string(result.MyPrivateKey))
	assert.Equal(t, expectedSharedSecret, string(result.SharedSecret))
}

func TestExtractPayloadFailsIfBodyIsInvalid(t *testing.T) {
	initEnv(envValues{
		privateKey:   b64UrlEncode("valid key"),
		sharedSecret: b64UrlEncode("valid secret"),
	})
	defer clearEnv()
	cfg.ParseConfig()

	req := events.LambdaFunctionURLRequest{
		IsBase64Encoded: true,
		Headers:         map[string]string{"encryption": b64UrlEncodeKeyVal("salt", "foobar"), "crypto-key": b64UrlEncodeKeyVal("dh", "foobar")},
		Body:            "**$(#", // invalid
	}
	_, err := http.ExtractPayload(req)
	assert.ErrorIs(t, err, http.ErrNotBase64Encoded)
	assert.ErrorContains(t, err, "[body decode failed]")
}

func TestExtractPayloadFailsIfPrivateKeyIsInvalid(t *testing.T) {
	initEnv(envValues{
		privateKey:   "$*))", // invalid
		sharedSecret: b64UrlEncode("valid"),
	})
	defer clearEnv()
	cfg.ParseConfig()

	req := events.LambdaFunctionURLRequest{
		IsBase64Encoded: true,
		Headers:         map[string]string{"encryption": b64UrlEncodeKeyVal("salt", "foobar"), "crypto-key": b64UrlEncodeKeyVal("dh", "foobar")},
	}
	_, err := http.ExtractPayload(req)
	assert.ErrorIs(t, err, http.ErrNotBase64Encoded)
	assert.ErrorContains(t, err, "[private key decode failed]")
}

func TestExtractPayloadFailsIfSharedSecretIsInvalid(t *testing.T) {
	initEnv(envValues{
		privateKey:   b64UrlEncode("valid"),
		sharedSecret: "$*((", // invalid
	})
	defer clearEnv()
	cfg.ParseConfig()

	req := events.LambdaFunctionURLRequest{
		IsBase64Encoded: true,
		Headers:         map[string]string{"encryption": b64UrlEncodeKeyVal("salt", "foobar"), "crypto-key": b64UrlEncodeKeyVal("dh", "foobar")},
	}
	_, err := http.ExtractPayload(req)
	assert.ErrorIs(t, err, http.ErrNotBase64Encoded)
	assert.ErrorContains(t, err, "[shared secret decode failed]")
}

func TestExtractPayloadFailsIfSaltIsInvalid(t *testing.T) {
	testCases := []struct {
		value       string
		expectedErr error
		desc        string
	}{
		{"", http.ErrMissingHeader, "header does not exist"},
		{"foobar", http.ErrInvalidHeader, "header is not a key/val pair"},
		{"salt=8#$*)#$", http.ErrNotBase64Encoded, "header val is not base64 encoded"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req := events.LambdaFunctionURLRequest{
				IsBase64Encoded: true,
				Headers:         map[string]string{"encryption": tc.value, "crypto-key": fmt.Sprintf("dh=%s", b64UrlEncode("foo"))},
			}
			_, err := http.ExtractPayload(req)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestExtractPayloadFailsIfEventIsNotBase64Encoded(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		IsBase64Encoded: false,
	}
	_, err := http.ExtractPayload(req)
	assert.ErrorIs(t, err, http.ErrNotBase64Encoded)
}

func TestExtractPayloadFailsIfTheirPublicKeyIsInvalid(t *testing.T) {
	testCases := []struct {
		value       string
		expectedErr error
		desc        string
	}{
		{"", http.ErrMissingHeader, "header does not exist"},
		{"foobar", http.ErrInvalidHeader, "header is not a key/val pair"},
		{"dh=8#$*)#$", http.ErrNotBase64Encoded, "header val is not base64 encoded"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req := events.LambdaFunctionURLRequest{
				IsBase64Encoded: true,
				Headers:         map[string]string{"crypto-key": tc.value},
			}
			_, err := http.ExtractPayload(req)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestExtractTargetsReturnsEmptySliceWhenNoPath(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		RawPath: "",
	}
	result, err := http.ExtractTargets(req)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestExtractTargetsReturnsErrorWhenPathIsNotB64Encoded(t *testing.T) {
	req := events.LambdaFunctionURLRequest{
		RawPath: "/&&($@/*#*)@",
	}
	_, err := http.ExtractTargets(req)
	assert.ErrorIs(t, err, http.ErrTargetDecode)
}

func TestExtractTargetsReturnsDecodedPathValues(t *testing.T) {
	encodedVal := base64.RawURLEncoding.EncodeToString([]byte("foobar"))
	testCases := []struct {
		input string
		count int
		desc  string
	}{
		{encodedVal, 1, "path of size 1; no trailing slash"},
		{fmt.Sprintf("%[1]s/%[1]s", encodedVal), 2, "path of size 2; no trailing slash"},
		{fmt.Sprintf("%[1]s/%[1]s/%[1]s", encodedVal), 3, "path of size 3; no trailing slash"},
		{fmt.Sprintf("%s/", encodedVal), 1, "path of size 1; trailing slash"},
		{fmt.Sprintf("%[1]s/%[1]s/", encodedVal), 2, "path of size 2; trailing slash"},
		{fmt.Sprintf("%[1]s/%[1]s/%[1]s/", encodedVal), 3, "path of size 3; trailing slash"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf(tc.desc, tc.count), func(t *testing.T) {
			req := events.LambdaFunctionURLRequest{
				RawPath: tc.input,
			}
			result, err := http.ExtractTargets(req)
			assert.Nil(t, err)
			assert.Equal(t, tc.count, len(result))
			assert.Equal(t, "foobar", result[0])
		})
	}
}

func TestEncodeResponseReturnsStatusCodeAndMessageReceived(t *testing.T) {
	input := map[string]string{"status": "ok"}
	expected, _ := json.Marshal(input)
	resp := http.EncodeResponse(200, "ok")
	assert.Equal(t, 200, resp.StatusCode)
	assert.JSONEq(t, string(expected), resp.Body)
}

func b64StdEncode(input string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(input))
}

func b64UrlEncode(input string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(input))
}

func b64UrlEncodeKeyVal(key string, val string) string {
	return fmt.Sprintf("%s=%s", key, base64.RawURLEncoding.EncodeToString([]byte(val)))
}

func b64UrlEncodeKeyValBytes(key string, val []byte) string {
	return fmt.Sprintf("%s=%s", key, base64.RawURLEncoding.EncodeToString(val))
}

type envValues struct {
	privateKey   string
	sharedSecret string
}

func initEnv(envVals envValues) {
	os.Setenv("MSTDN_PRIVATE_KEY", envVals.privateKey)
	os.Setenv("MSTDN_SHARED_SECRET", envVals.sharedSecret)
}

func clearEnv() {
	for _, kv := range os.Environ() {
		data := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(data[0], "MSTDN_") {
			os.Unsetenv(data[0])
		}
	}
}

func genP256PublicKey() *ecdsa.PublicKey {
	buf := make([]byte, 128)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(buf))
	if err != nil {
		panic(err)
	}
	return &key.PublicKey
}
