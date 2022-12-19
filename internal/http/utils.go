package http

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
	"crypto/ecdsa"
	"crypto/elliptic"
	b64 "encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const authHeaderName = "authorization"
const crytoKeyHeaderName = "crypto-key"
const encryptionHeaderName = "encryption"

const p256CryptoKeyID = "p256ecdsa"
const dhCryptoKeyID = "dh"
const saltKeyID = "salt"

const authTokenPrefix = "WebPush "

func parseSalt(event events.LambdaFunctionURLRequest) ([]byte, error) {
	hdr := event.Headers[encryptionHeaderName]
	if hdr == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingHeader, encryptionHeaderName)
	}

	keys, err := parseKeyValHeader(hdr)
	if err != nil {
		return nil, err
	}

	decode, err := b64.RawURLEncoding.DecodeString(keys[saltKeyID])
	if err != nil {
		err = fmt.Errorf("[salt decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
	}
	return decode, err
}

func parseTheirPublicKey(event events.LambdaFunctionURLRequest) ([]byte, error) {
	hdr := event.Headers[crytoKeyHeaderName]
	if hdr == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingHeader, crytoKeyHeaderName)
	}

	keys, err := parseKeyValHeader(hdr)
	if err != nil {
		return nil, err
	}

	decode, err := b64.RawURLEncoding.DecodeString(keys[dhCryptoKeyID])
	if err != nil {
		err = fmt.Errorf("[their public key decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
	}
	return decode, err
}

func extractBearerToken(event events.LambdaFunctionURLRequest) (string, error) {
	authHeader := event.Headers[authHeaderName]
	if authHeader == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingHeader, authHeaderName)
	}

	if !strings.HasPrefix(authHeader, authTokenPrefix) {
		return "", fmt.Errorf("%w: %s", ErrInvalidHeader, authHeaderName)
	}

	return authHeader[len(authTokenPrefix):], nil
}

func parseP256PublicKey(event events.LambdaFunctionURLRequest) (*ecdsa.PublicKey, error) {
	hdr := event.Headers[crytoKeyHeaderName]
	if hdr == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingHeader, crytoKeyHeaderName)
	}

	cryptoKeys, err := parseKeyValHeader(hdr)
	if err != nil {
		return nil, err
	}

	return b64ToPublicKey(cryptoKeys[p256CryptoKeyID])
}

func b64ToPublicKey(input string) (*ecdsa.PublicKey, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "public key cannot be empty")
	}

	p256Key, err := b64.RawURLEncoding.DecodeString(input)
	if err != nil {
		e := fmt.Errorf("[public key decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
		return nil, e
	}

	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, p256Key)
	if x == nil {
		return nil, fmt.Errorf("%s: %w", "elliptic unmarshal failed", ErrCryptoFailure)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

func parseKeyValHeader(hdrVal string) (map[string]string, error) {
	result := make(map[string]string)
	keys := strings.Split(hdrVal, ";")
	for _, pair := range keys {
		v := strings.Split(pair, "=")
		if len(v) != 2 {
			return nil, fmt.Errorf("%w: invalid key/val header: %v", ErrInvalidHeader, v)
		}
		result[v[0]] = v[1]
	}
	return result, nil
}
