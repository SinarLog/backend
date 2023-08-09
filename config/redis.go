package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type redisConfig struct {
	Address  string
	Password string
	Db       int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	MinIdleConn int
	MaxIdleConn int
	MaxIdleTime time.Duration
}

func (c *Config) newRedisConfig() {
	r := redisConfig{
		Address:  fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Unable to parse redis db %s\n", err)
	}
	r.Db = db

	if x := os.Getenv("REDIS_READ_TIMEOUT"); x != "" {
		readTimeout, err := time.ParseDuration(x)
		if err != nil {
			log.Fatalf("Unable to parse redis readtimeout %s\n", err)
		}
		r.ReadTimeout = readTimeout
	}

	if x := os.Getenv("REDIS_WRITE_TIMEOUT"); x != "" {
		writeTimout, err := time.ParseDuration(x)
		if err != nil {
			log.Fatalf("Unable to parse redis writetimeout %s\n", err)
		}
		r.WriteTimeout = writeTimout
	}

	if x := os.Getenv("REDIS_MIN_IDLE_CONN"); x != "" {
		minIdleConn, err := strconv.Atoi(x)
		if err != nil {
			log.Fatalf("Unable to parse redis min idle conn %s\n", err)
		}
		r.MinIdleConn = minIdleConn
	}

	if x := os.Getenv("REDIS_MAX_IDLE_CONN"); x != "" {
		maxIdleConn, err := strconv.Atoi(x)
		if err != nil {
			log.Fatalf("Unable to parse redis max idle conn %s\n", err)
		}
		r.MaxIdleConn = maxIdleConn
	}

	if x := os.Getenv("REDIS_MAX_IDLE_TIME"); x != "" {
		maxIdleTime, err := time.ParseDuration(x)
		if err != nil {
			log.Fatalf("Unable to parse redis max idle time %s\n", err)
		}
		r.MaxIdleTime = maxIdleTime
	}

	if err := r.validate(); err != nil {
		log.Fatalf("FATAL - %s", err)
	}

	c.Redis = r
}

func (r redisConfig) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Password, validation.Required.Error("(redisConfig).validate: redis password is required for security reason")),
		validation.Field(&r.Db, validation.Max(15).Error("(redisConfig.validate: invalid redis db index)"), validation.Min(0).Error("(redisConfig.validate: invalid redis db index)")),
	)
}
