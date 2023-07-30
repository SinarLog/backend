package service

import (
	"context"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
	"sinarlog.com/pkg/doorkeeper"
)

type doorkeeperService struct {
	dk *doorkeeper.Doorkeeper
}

func NewDoorkeeperService(dk *doorkeeper.Doorkeeper) *doorkeeperService {
	return &doorkeeperService{dk}
}

/*
---------- Hashing Section ----------
*/
func (s *doorkeeperService) HashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}

func (s *doorkeeperService) VerifyPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

/*
---------- JWT Section ----------
*/
func (s *doorkeeperService) GenerateToken(employee entity.Employee) (token string, err error) {
	now := time.Now().In(utils.CURRENT_LOC)
	claims := jwt.MapClaims{
		"iss": s.dk.GetIssuer(),
		"eat": now.Add(s.dk.AccessDuration).Unix(),
		"iat": now.Unix(),
		"id":  employee.Id,
		"nbf": now.Unix(),
	}

	return jwt.NewWithClaims(s.dk.GetSignMethod(), claims).SignedString(s.dk.GetPrivKey())
}

func (s *doorkeeperService) VerifyAndParseToken(ctx context.Context, tk string) (string, error) {
	claims, err := s.verifyAndGetClaims(tk)
	if err != nil {
		return "", err
	}

	if err := s.verifyClaims(ctx, claims, "id"); err != nil {
		return "", err
	}

	return claims["id"].(string), nil
}

func (s *doorkeeperService) verifyAndGetClaims(tk string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tk, func(t *jwt.Token) (interface{}, error) {
		switch s.dk.GetConcreteSignMethod() {
		case doorkeeper.RSA_SIGN_METHOD_TYPE:
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("signing method invalid")
			}
		case doorkeeper.RSAPSS_SIGN_METHOD_TYPE:
			if _, ok := t.Method.(*jwt.SigningMethodRSAPSS); !ok {
				return nil, fmt.Errorf("signing method invalid")
			}
		case doorkeeper.HMAC_SIGN_METHOD_TYPE:
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method invalid")
			}
		case doorkeeper.ECDSA_SIGN_METHOD_TYPE:
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("signing method invalid")
			}
		case doorkeeper.EdDSA_SIGN_METHOD_TYPE:
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("signing method invalid")
			}
		}
		return s.dk.GetPubKey(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("validate: invalid")
	}

	return claims, nil
}

func (s *doorkeeperService) verifyClaims(ctx context.Context, claims jwt.MapClaims, expectedKeys ...any) error {
	keys := []any{"iss", "iat", "eat", "nbf"}
	keys = append(keys, []any(expectedKeys)...)

	if err := s.validateKeys(ctx, claims, keys...); err != nil {
		return err
	}

	now := time.Now().In(utils.CURRENT_LOC).UTC()

	if _, ok := claims["iss"].(string); !ok {
		return fmt.Errorf("invalid token claims")
	}
	if _, ok := claims["eat"].(float64); !ok {
		return fmt.Errorf("invalid token claims")
	}
	if _, ok := claims["nbf"].(float64); !ok {
		return fmt.Errorf("invalid token claims")
	}

	if now.Unix() > int64(claims["eat"].(float64)) {
		return fmt.Errorf("token has expired")
	}

	if int64(claims["nbf"].(float64)) > now.Unix() {
		return fmt.Errorf("invalid token claims: nbf > now")
	}

	if claims["iss"].(string) != s.dk.GetIssuer() {
		return fmt.Errorf("unrecognized issuer")
	}

	return nil
}

func (s *doorkeeperService) validateKeys(ctx context.Context, obj map[string]any, args ...any) error {
	keys := make([]string, len(obj))

	index := 0
	for k := range obj {
		keys[index] = k
		index++
	}

	return validation.ValidateWithContext(ctx, keys, validation.Each(validation.Required, validation.In(args...).Error("does not contained required claim")))
}

/*
---------- OTP Section ----------
*/
func (s *doorkeeperService) GenerateOTP() (string, int64, time.Duration) {
	now := time.Now().In(utils.CURRENT_LOC).Unix()
	otp := s.dk.GetTOTP().At(now)

	return otp, now, s.dk.GetOTPExpDuration()
}

func (s *doorkeeperService) VerifyOTP(otp string, timestamp int64) bool {
	return s.dk.GetTOTP().Verify(otp, timestamp)
}
