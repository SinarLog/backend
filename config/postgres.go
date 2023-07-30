package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type dbConfig struct {
	URL      string
	ExecURL  string
	host     string
	port     string
	user     string
	password string
	name     string

	MaxPoolSize     int
	MaxOpenConn     int
	MaxConnLifetime time.Duration
}

// newDbConfig method    has a receiver of the config
// struct. It loads the dbConfig struct into the main
// Config struct.
func (c *Config) newDbConfig() {
	d := dbConfig{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		name:     os.Getenv("DB_NAME"),
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
	}

	maxPoolSize, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONN"))
	if err != nil {
		log.Fatalf("Unable to parse postgres pool size %s\n", err)
	}
	d.MaxPoolSize = maxPoolSize

	maxOpenConn, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	if err != nil {
		log.Fatalf("Unable to parse postgres open conn %s\n", err)
	}
	d.MaxOpenConn = maxOpenConn

	maxConnLifetime, err := time.ParseDuration(os.Getenv("DB_MAX_CONN_LIFETIME"))
	if err != nil {
		log.Fatalf("Unable to parse postgres conn lifetime %s\n", err)
	}
	d.MaxConnLifetime = maxConnLifetime

	if err := d.validate(); err != nil {
		log.Fatalf("%s", err)
	}

	c.Db = d

	// Create dsn
	dsn := fmt.Sprintf("postgres://%s:%s/%s", d.host, d.port, d.name)
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("ERROR parsing dsn: %s\n", err)
	}
	u.User = url.UserPassword(d.user, d.password)

	c.Db.URL = u.String()
}

// validate method    validates the dbConfig struct
// such that in matches the expected conditions.
func (d dbConfig) validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.host, validation.Required, is.Host),
		validation.Field(&d.port, validation.Required, is.Port.Error("hello world")),
		validation.Field(&d.user, validation.Required),
		validation.Field(&d.name, validation.Required),
		validation.Field(&d.password, validation.Required),
		validation.Field(&d.MaxConnLifetime, validation.Required),
		validation.Field(&d.MaxOpenConn, validation.Required),
		validation.Field(&d.MaxPoolSize, validation.Required),
	)
}
