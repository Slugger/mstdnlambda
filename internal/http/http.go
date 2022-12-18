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
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"github.com/slugger/mstdnlambda/internal/cfg"
	"github.com/slugger/mstdnlambda/internal/jwt"
	"github.com/slugger/mstdnlambda/internal/payload"
)

var ErrInvalidHeader = errors.New("header contains invalid contents")
var ErrMissingHeader = errors.New("expected header not found")
var ErrInvalidInput = errors.New("invalid input received")
var ErrCryptoFailure = errors.New("crypto failure")
var ErrNotBase64Encoded = errors.New("input is not base64 encoded")
var ErrTargetDecode = errors.New("target decode failed")

func ExtractJwt(event events.LambdaFunctionURLRequest) (*jwt.VerifiableJwt, error) {
	var token, aud string
	var err error

	if token, err = extractBearerToken(event); err != nil {
		return nil, err
	}

	aud = fmt.Sprintf("https://%s", event.RequestContext.DomainName)

	publicKey, err := parseP256PublicKey(event)
	if err != nil {
		return nil, err
	}

	return &jwt.VerifiableJwt{
		Token:     token,
		Aud:       aud,
		PublicKey: publicKey,
	}, nil
}

func ExtractPayload(event events.LambdaFunctionURLRequest) (*payload.EncryptedPayload, error) {
	if !event.IsBase64Encoded {
		return nil, ErrNotBase64Encoded // AWS will not send raw binary streams to us
	}

	var sharedSecret, myPrivateKey, theirPublicKey, salt, data []byte
	var err error

	if theirPublicKey, err = parseTheirPublicKey(event); err != nil {
		return nil, err
	}

	if salt, err = parseSalt(event); err != nil {
		return nil, err
	}

	if sharedSecret, err = b64.RawURLEncoding.DecodeString(cfg.Cfg.SharedSecret()); err != nil {
		e := fmt.Errorf("[shared secret decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
		return nil, e
	}

	if myPrivateKey, err = b64.RawURLEncoding.DecodeString(cfg.Cfg.PrivateKey()); err != nil {
		e := fmt.Errorf("[private key decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
		return nil, e
	}

	if data, err = b64.StdEncoding.DecodeString(event.Body); err != nil {
		e := fmt.Errorf("[body decode failed] %w: %s", ErrNotBase64Encoded, err.Error())
		return nil, e
	}

	return &payload.EncryptedPayload{
		SharedSecret:   sharedSecret,
		MyPrivateKey:   myPrivateKey,
		TheirPublicKey: theirPublicKey,
		Salt:           salt,
		Data:           data,
	}, nil
}

func ExtractTargets(event events.LambdaFunctionURLRequest) ([]string, error) {
	encodedTargets := strings.Split(event.RawPath, "/")
	targets := make([]string, 0)
	for _, t := range encodedTargets {
		if t == "" {
			continue
		}
		s, err := b64.RawURLEncoding.DecodeString(t)
		if err != nil {
			return targets, errors.Wrap(ErrTargetDecode, err.Error())
		}
		targets = append(targets, string(s))
	}
	return targets, nil
}

func EncodeResponse(code int, msg string) *events.LambdaFunctionURLResponse {
	resp := map[string]string{
		"status": msg,
	}
	enc, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	return &events.LambdaFunctionURLResponse{
		StatusCode:      code,
		Body:            string(enc),
		IsBase64Encoded: false,
	}
}
