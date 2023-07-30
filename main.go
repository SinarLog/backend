package main

import (
	"sinarlog.com/cmd/app"
	"sinarlog.com/config"
)

func main() {
	cfg := config.GetConfig()

	app.Run(cfg)
}
