package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type doorkeeperConfig struct {
	// --- JWT Config ---
	AccessDuration  time.Duration
	RefreshDuration time.Duration

	SigningMethod string
	SignSize      string
	PubPath       string
	PrivPath      string
	Issuer        string

	// --- Password Hasher Config ---
	HashMethod string

	// --- OTP Config ---
	OTPExp          time.Duration
	OTPSecretLength int
}

// newServerConfig method    has a Config receiver
// such that it loads the serverConfig to the main
// Config.
func (c *Config) newDoorkeeperConfig() {
	d := doorkeeperConfig{
		SigningMethod: strings.ToUpper(os.Getenv("DOORKEEPER_SIGNING_METHOD")),
		SignSize:      strings.ToLower(os.Getenv("DOORKEEPER_SIGN_SIZE")),
		PrivPath:      strings.ToLower(os.Getenv("DOORKEEPER_CERT_PRIVATE_PATH")),
		PubPath:       strings.ToLower(os.Getenv("DOORKEEPER_CERT_PUBLIC_PATH")),
		Issuer:        strings.ToUpper(os.Getenv("DOORKEEPER_ISSUER")),
		HashMethod:    strings.ToUpper(os.Getenv("DOORKEEPER_HASH_METHOD")),
	}

	otpExp, err := time.ParseDuration(strings.ToLower(os.Getenv("DOORKEEPER_OTP_EXPIRATION_DURATION")))
	if err != nil {
		log.Fatalf("Error while parsing otp expiration %s", err)
	}
	d.OTPExp = otpExp

	accessDuration, err := time.ParseDuration(strings.ToLower(os.Getenv("DOORKEEPER_ACCESS_TOKEN_DURATION")))
	if err != nil {
		log.Fatalf("Error while parsing access token duration %s", err)
	}
	d.AccessDuration = accessDuration

	refreshDuration, err := time.ParseDuration(strings.ToLower(os.Getenv("DOORKEEPER_REFRESH_TOKEN_DURATION")))
	if err != nil {
		log.Fatalf("Error while parsing refresh token duration %s", err)
	}
	d.RefreshDuration = refreshDuration

	otpSecretLength, err := strconv.Atoi(os.Getenv("DOORKEEPER_OTP_SECRET_LENGTH"))
	if err != nil {
		log.Fatalf("Error while parsing otp secret length %s", err)
	}
	d.OTPSecretLength = otpSecretLength

	if err := d.validate(); err != nil {
		log.Fatalf("FATAL - %s", err)
	}

	c.Doorkeeper = d
}

// validate method    validates the serverConfig
// values such that it meets the requirements.
func (d doorkeeperConfig) validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.SigningMethod, validation.Required.
			Error("Please provide a signing method in the environment. This is needed for signing authorization tokens"),
			validation.In("HMAC", "RSA", "ECDSA", "RSA-PSS", "EdDSA")),
		validation.Field(&d.SignSize,
			validation.When(d.SigningMethod != "EdDSA", validation.Required.
				Error("Please provide a signing size in the environment. This is needed for signing authorization tokens"),
				validation.In("256", "384", "512"))),
		validation.Field(&d.HashMethod, validation.Required.
			Error("Please provide hash method in the environment. This is needed when hashing credentials"),
			validation.In("SHA1", "SHA224", "SHA256",
				"SHA384", "SHA512", "SHA3_224", "SHA3_256", "SHA3_384", "SHA3_512")),
		validation.Field(&d.OTPExp, validation.Required.Error("Please provide a OTP expiration duration"),
			validation.By(validateEmptyDuration)),
		validation.Field(&d.AccessDuration, validation.Required, validation.By(validateEmptyDuration)),
		validation.Field(&d.RefreshDuration, validation.Required, validation.By(validateEmptyDuration)),
		validation.Field(&d.PrivPath, validation.Required.Error("Please provide a private key as it is required for encryption")),
	)
}
