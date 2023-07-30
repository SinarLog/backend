package vo

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Credential struct {
	Email    string
	Password string

	OTP string

	OldPassword          string
	NewPassword          string
	ConfirmationPassword string

	AccessToken  string
	RefreshToken string
}

func (v Credential) ValidateAuthentication() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Email, validation.Required.Error("email field is required"),
		),
		validation.Field(&v.Password, validation.Required.Error("password field is required"),
			validation.Length(6, 0).Error("minimum length of password must be 8"),
		),
	)
}
