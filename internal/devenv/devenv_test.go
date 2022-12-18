package devenv

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapString(t *testing.T) {
	sut := envMap{}
	sut["foo"] = "bar"
	assert.Equal(t, "[foo=bar]", sut.String())
}

func TestEnvMapSet(t *testing.T) {
	sut := envMap{}
	err := sut.Set("foo=bar")
	assert.Nil(t, err)
	assert.Equal(t, "[foo=bar]", sut.String())
}

func TestEnvMapSetFailsWithInvalidInput(t *testing.T) {
	sut := envMap{}
	assert.NotNil(t, sut.Set("foobar"))
}
