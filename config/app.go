package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type appConfig struct {
	Environment             string
	LogPath                 string
	RaterLimit              int
	BurstLimit              int
	RaterEvaluationInterval time.Duration
	RaterDeletionTime       time.Duration
	DefaultPaginationSize   int
	MailerEmailAddress      string
	MailerEmailPassword     string
	MailerTemplatePath      string

	// Google Cloud Related
	GoogleProjectId          string
	GoogleServiceAccountPath string
}

// newServerConfig method    has a Config receiver
// such that it loads the serverConfig to the main
// Config.
func (c *Config) newAppConfig() {
	a := appConfig{
		Environment:         strings.ToUpper(os.Getenv("GO_ENV")),
		LogPath:             strings.ToLower(os.Getenv("LOG_PATH")),
		MailerEmailAddress:  strings.ToLower(os.Getenv("MAILER_SENDER_ADDRESS")),
		MailerTemplatePath:  strings.ToLower(os.Getenv("MAILER_TEMPLATE_PATH")),
		MailerEmailPassword: os.Getenv("MAILER_SENDER_PASSWORD"),

		GoogleProjectId:          os.Getenv("GOOGLE_PROJECT_ID"),
		GoogleServiceAccountPath: strings.ToLower(os.Getenv("GOOGLE_KEY_PATH")),
	}

	defaultPaginationSize, err := strconv.Atoi(os.Getenv("DEFAULT_ROWS_PER_PAGE"))
	if err != nil {
		log.Fatalf("Unable to parse app pagination %s\n", err)
	}
	a.DefaultPaginationSize = defaultPaginationSize

	raterEvInt, err := time.ParseDuration(os.Getenv("RATER_EVALUATION_INTERVAL"))
	if err != nil {
		log.Fatalf("Unable to parse app rate eval time %s\n", err)
	}
	a.RaterEvaluationInterval = raterEvInt

	raterDelTime, err := time.ParseDuration(os.Getenv("RATER_DELETION_TIME"))
	if err != nil {
		log.Fatalf("Unable to parse rater del time %s\n", err)
	}
	a.RaterDeletionTime = raterDelTime

	raterLimit, err := strconv.Atoi(os.Getenv("RATER_LIMIT"))
	if err != nil {
		log.Fatalf("Unable to parse rater limit %s\n", err)
	}
	a.RaterLimit = raterLimit

	burstLimit, err := strconv.Atoi(os.Getenv("BURST_LIMIT"))
	if err != nil {
		log.Fatalf("Unable to parse burst limit %s\n", err)
	}
	a.BurstLimit = burstLimit

	if err := a.validate(); err != nil {
		log.Fatalf("FATAL - %s", err)
	}

	c.App = a
}

// validate method    validates the serverConfig
// values such that it meets the requirements.
func (a appConfig) validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.DefaultPaginationSize, validation.Required, validation.Min(5)),
		validation.Field(&a.LogPath, validation.Required),
		validation.Field(&a.RaterDeletionTime, validation.Required, validation.By(validateEmptyDuration)),
		validation.Field(&a.RaterEvaluationInterval, validation.Required, validation.By(validateEmptyDuration)),
		validation.Field(&a.LogPath, validation.Required),
		validation.Field(&a.Environment, validation.Required, validation.In(
			PRODUCTION,
			STAGING,
			TESTING,
			DEVELOPMENT,
		)),
		validation.Field(&a.MailerEmailAddress, validation.Required, is.Email),
		validation.Field(&a.MailerEmailPassword, validation.Required),
	)
}
