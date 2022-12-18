package jwt

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
