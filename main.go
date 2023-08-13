package main

import (
	"log"
	"os"

	"sinarlog.com/cmd/app"
	"sinarlog.com/config"
)

func main() {
	if os.Getenv("GO_ENV") == "" {
		if err := os.Setenv("GO_ENV", "DEVELOPMENT"); err != nil {
			log.Fatalf("unable to set GO_ENV to DEVELOPMENT: %s\n", err)
		}
	}
	cfg := config.GetConfig()

	app.Run(cfg)
}
