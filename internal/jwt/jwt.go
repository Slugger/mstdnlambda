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

// ErrInvalidJwtSigningMethod represents an error that denotes a JWT token is not signed with a supported encryption method
var ErrInvalidJwtSigningMethod = errors.New("unexpected signing method")

// ErrJwtParseFailure represents an error denoting that a JWT token could not be parsed properly
var ErrJwtParseFailure = errors.New("jwt parse failed")

// Verify parses and verifies the given JWT token; returns an error iff the token validation failed otherwise returns nil
func Verify(vjwt *VerifiableJwt) error {
	if _, err := jwtlib.Parse(vjwt.Token, vjwt.publicKey, jwtlib.WithAudience(vjwt.Aud)); err != nil {
		return fmt.Errorf("%w: %s", ErrJwtParseFailure, err.Error())
	}
	return nil
}

// VerifiableJwt A dot encoded JWT token, its expected audience and the ECDSA public key needed to verify the token
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
