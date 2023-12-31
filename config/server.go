package config

import (
	"log"
	"os"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type serverConfig struct {
	Host string
	Port string
}

// newServerConfig method    has a Config receiver
// such that it loads the serverConfig to the main
// Config.
func (c *Config) newServerConfig() {
	s := serverConfig{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}

	if err := s.validate(); err != nil {
		log.Fatalf("%s", err)
	}

	c.Server = s
}

// validate method    validates the serverConfig
// values such that it meets the requirements.
func (d serverConfig) validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Host, validation.When(d.Host != "", validation.Required, is.Host.Error("(serverConfig).validate: unrecognised host for server"))),
		validation.Field(&d.Port, validation.Required, is.Port.Error("(serverConfig).validate: unrecognised port for server")),
	)
}
