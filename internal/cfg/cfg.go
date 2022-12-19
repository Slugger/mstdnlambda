package cfg

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

import "github.com/caarlos0/env/v6"

// Config represents all config options for the lambda
type Config interface {
	AwsRegion() string
	LogLevel() string
	PrivateKey() string
	SharedSecret() string
	IsSkipJwtVerify() bool
	IsSkipPayloadDecrypt() bool
}

// Cfg is the global Config instance for the lambda
var Cfg Config

func init() {
	Cfg = &configSettings{}
}

// NewConfig returns the default Config instance
func NewConfig() Config {
	return Cfg
}

// ParseConfig parses the env variables and configures the lambda
func ParseConfig() Config {
	cfg := configSettings{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	Cfg = &cfg
	return Cfg
}

type configSettings struct {
	AwsRegionValue     string `env:"MSTDN_AWS_REGION" envDefault:"ca-central-1"`
	LogLevelValue      string `env:"MSTDN_LOG_LEVEL" envDefault:"INFO"`
	PrivateKeyValue    string `env:"MSTDN_PRIVATE_KEY,notEmpty,unset"`
	SharedSecretValue  string `env:"MSTDN_SHARED_SECRET,notEmpty,unset"`
	SkipJwtVerify      bool   `env:"MSTDN_SKIP_JWT_VERIFY" envDefault:"false"`
	SkipPayloadDecrypt bool   `env:"MSTDN_SKIP_PAYLOAD_DECRYPT" envDefault:"false"`
}

func (c *configSettings) AwsRegion() string          { return c.AwsRegionValue }
func (c *configSettings) IsSkipJwtVerify() bool      { return c.SkipJwtVerify }
func (c *configSettings) IsSkipPayloadDecrypt() bool { return c.SkipPayloadDecrypt }
func (c *configSettings) LogLevel() string           { return c.LogLevelValue }
func (c *configSettings) PrivateKey() string         { return c.PrivateKeyValue }
func (c *configSettings) SharedSecret() string       { return c.SharedSecretValue }
