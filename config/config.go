package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
	"unicode"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     serverConfig
	Db         dbConfig
	Mongo      mongoConfig
	App        appConfig
	Doorkeeper doorkeeperConfig
	Redis      redisConfig
	Bucket     bucketConfig
}

var (
	once              sync.Once
	cfgSingleInstance *Config
)

const (
	PRODUCTION  string = "PRODUCTION"
	STAGING     string = "STAGING"
	TESTING     string = "TESTING"
	DEVELOPMENT string = "DEVELOPMENT"
)

// GetConfig function    either returns an already created
// config instance or creates a new config instance if there
// is none existing yet.
func GetConfig() *Config {
	if cfgSingleInstance == nil {
		once.Do(
			func() {
				log.Println("Creating a config single instance")
				cfgSingleInstance = new(Config)
				cfgSingleInstance.load()
				printInfo("", cfgSingleInstance)
			})
	}
	return cfgSingleInstance
}

// load method    loads other specific configurations. It should
// only be called during the creational of a new config instance.
// It is also the place to register your configs.
func (c *Config) load() {
	// All environments during production or testing must be
	// injected before the start of the application.
	if os.Getenv("GO_ENV") == "DEVELOPMENT" {
		// Change the file names as you will
		err := godotenv.Load(".env.development")
		if err != nil {
			log.Fatalf("unable to load environment file: %s", err)
		}
	}

	// Register here for your new configs with third-parties.
	c.newServerConfig()
	c.newDbConfig()
	c.newMongoConfig()
	c.newAppConfig()
	c.newDoorkeeperConfig()
	c.newRedisConfig()
	c.newBucketConfig()
}

// printInfo function    prints the entire configuration info
// to the terminal. This helps the developer to check whether
// the right configurations has been read upon starting the
// app.
func printInfo(indent string, abs any) {
	var t reflect.Type
	var v reflect.Value

	if reflect.TypeOf(abs).Kind() == reflect.Pointer {
		t = reflect.TypeOf(abs).Elem()
		v = reflect.ValueOf(abs).Elem()
	} else {
		t = reflect.TypeOf(abs)
		v = reflect.ValueOf(abs)
	}

	for i := 0; i < t.NumField(); i++ {
		if unicode.IsUpper(rune(t.Field(i).Name[0])) {
			if v.Field(i).Kind() == reflect.Struct {
				log.Printf("%s%s:\n", indent, t.Field(i).Name)
				printInfo("   ", v.Field(i).Interface())
			} else {
				log.Printf("%s%s: %v\n", indent, t.Field(i).Name, v.Field(i).Interface())
			}
		}
	}
}

func validateEmptyDuration(value any) error {
	if value != time.Duration(0) {
		return nil
	}
	return fmt.Errorf("time duration must not be wmpty")
}
