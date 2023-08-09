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

type mongoConfig struct {
	host     string
	port     string
	user     string
	password string

	DbName          string
	URI             string
	MaxPoolSize     int
	MaxOpenConn     int
	MaxConnLifetime time.Duration
}

// newDbConfig method    has a receiver of the config
// struct. It loads the dbConfig struct into the main
// Config struct.
func (c *Config) newMongoConfig() {
	d := mongoConfig{
		host:     os.Getenv("MONGO_HOST"),
		port:     os.Getenv("MONGO_PORT"),
		user:     os.Getenv("MONGO_USER"),
		password: os.Getenv("MONGO_PASSWORD"),
		DbName:   os.Getenv("MONGO_NAME"),
	}

	maxPoolSize, err := strconv.Atoi(os.Getenv("MONGO_MAX_OPEN_CONN"))
	if err != nil {
		log.Fatalf("Unable to parse mongo pool size %s\n", err)
	}
	d.MaxPoolSize = maxPoolSize

	maxOpenConn, err := strconv.Atoi(os.Getenv("MONGO_MAX_POOL_SIZE"))
	if err != nil {
		log.Fatalf("Unable to parse mongo open conn %s\n", err)
	}
	d.MaxOpenConn = maxOpenConn

	maxConnLifetime, err := time.ParseDuration(os.Getenv("MONGO_MAX_CONN_LIFETIME"))
	if err != nil {
		log.Fatalf("Unable to parse mongo conn lifetime %s\n", err)
	}
	d.MaxConnLifetime = maxConnLifetime

	if err := d.validate(); err != nil {
		log.Fatalf("%s", err)
	}

	c.Mongo = d

	// Create dsn
	dsn := fmt.Sprintf("mongodb://%s:%s/%s?authSource=admin", d.host, d.port, d.DbName)
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("ERROR parsing dsn: %s\n", err)
	}
	u.User = url.UserPassword(d.user, d.password)

	c.Mongo.URI = u.String()
}

// validate method    validates the dbConfig struct
// such that in matches the expected conditions.
func (d mongoConfig) validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.host, validation.Required, is.Host),
		validation.Field(&d.port, validation.Required, is.Port.Error("unrecognised port for mongo")),
		validation.Field(&d.user, validation.Required),
		validation.Field(&d.password, validation.Required),
		validation.Field(&d.MaxConnLifetime, validation.Required),
		validation.Field(&d.DbName, validation.Required.Error("please provide a mongo db name")),
		validation.Field(&d.MaxOpenConn, validation.Required),
		validation.Field(&d.MaxPoolSize, validation.Required),
	)
}
