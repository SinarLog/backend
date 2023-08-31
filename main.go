package main

import (
	"log"
	"os"
	"os/exec"

	"sinarlog.com/cmd/app"
	"sinarlog.com/config"
)

func main() {
	if os.Getenv("GO_ENV") == "" {
		if err := os.Setenv("GO_ENV", "DEVELOPMENT"); err != nil {
			log.Fatalf("unable to set GO_ENV to DEVELOPMENT: %s\n", err)
		}
	}

	gitCommitCmd := exec.Command("git", "rev-parse", "HEAD")
	gitCommit, err := gitCommitCmd.Output()
	if err != nil {
		log.Fatalf("error running git rev-parse: %s\n", err)
	}
	gitCommitStr := string(gitCommit)
	gitCommitStr = gitCommitStr[:len(gitCommitStr)-1]

	os.Setenv("GIT_COMMIT", gitCommitStr)

	cfg := config.GetConfig()

	app.Run(cfg)
}
