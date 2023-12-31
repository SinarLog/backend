package doorkeeper

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/sha3"
)

// Option -.
type Option func(*Doorkeeper)

func RegisterPrivatePath(path string) Option {
	return func(d *Doorkeeper) {
		d.privPath = path
	}
}

func RegisterPublicPath(path string) Option {
	return func(d *Doorkeeper) {
		d.pubPath = path
	}
}

func RegisterAccessDuration(t time.Duration) Option {
	return func(d *Doorkeeper) {
		if t > 0 {
			d.AccessDuration = t
		}
	}
}

func RegisterRefreshDuration(t time.Duration) Option {
	return func(d *Doorkeeper) {
		if t > 0 {
			d.RefreshDuration = t
		}
	}
}

func RegisterIssuer(iss string) Option {
	return func(d *Doorkeeper) {
		d.issuer = iss
	}
}

func RegisterOTPSecretLength(val int) Option {
	return func(d *Doorkeeper) {
		if val > 0 {
			d.otpSecretLength = val
		}
	}
}

func RegisterOTPExpDuration(t time.Duration) Option {
	return func(d *Doorkeeper) {
		if t > 0 {
			d.otpExpDuration = t
		}
	}
}

func RegisterHasherFunc(alg string) Option {
	return func(d *Doorkeeper) {
		switch alg {
		case "SHA1":
			d.hasherFunc = sha1.New
		case "SHA224":
			d.hasherFunc = sha256.New224
		case "SHA256":
			d.hasherFunc = sha256.New
		case "SHA384":
			d.hasherFunc = sha512.New384
		case "SHA512":
			d.hasherFunc = sha512.New
		case "SHA3_224":
			d.hasherFunc = sha3.New224
		case "SHA3_256":
			d.hasherFunc = sha3.New256
		case "SHA3_384":
			d.hasherFunc = sha3.New384
		case "SHA3_512":
			d.hasherFunc = sha3.New512
		}
	}
}

func RegisterSignMethod(alg, size string) Option {
	return func(d *Doorkeeper) {
		switch alg {
		case "HMAC":
			switch size {
			case "256":
				d.signMethod = jwt.SigningMethodHS256
			case "384":
				d.signMethod = jwt.SigningMethodHS384
			case "512":
				d.signMethod = jwt.SigningMethodHS512
			}
		case "RSA":
			switch size {
			case "256":
				d.signMethod = jwt.SigningMethodRS256
			case "384":
				d.signMethod = jwt.SigningMethodRS384
			case "512":
				d.signMethod = jwt.SigningMethodRS512
			}
		case "ECDSA":
			switch size {
			case "256":
				d.signMethod = jwt.SigningMethodES256
			case "384":
				d.signMethod = jwt.SigningMethodES384
			case "512":
				d.signMethod = jwt.SigningMethodES512
			}
		case "RSA-PSS":
			switch size {
			case "256":
				d.signMethod = jwt.SigningMethodPS256
			case "384":
				d.signMethod = jwt.SigningMethodPS384
			case "512":
				d.signMethod = jwt.SigningMethodPS512
			}
		case "EdDSA":
			d.signMethod = &jwt.SigningMethodEd25519{}
		}
	}
}
