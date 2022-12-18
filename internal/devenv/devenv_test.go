package devenv

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
