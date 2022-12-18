package cfg

import "github.com/caarlos0/env/v6"

type Config interface {
	AwsRegion() string
	LogLevel() string
	PrivateKey() string
	SharedSecret() string
	IsSkipJwtVerify() bool
	IsSkipPayloadDecrypt() bool
}

var Cfg Config

func init() {
	Cfg = &configSettings{}
}

func NewConfig() Config {
	return Cfg
}

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
