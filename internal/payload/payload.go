package payload

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
