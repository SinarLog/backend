package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID           string       `json:"id,omitempty"`
	Email        string       `json:"email,omitempty"`
	FullName     string       `json:"fullName,omitempty"`
	Avatar       string       `json:"avatar"`
	IsNewUser    bool         `json:"isNewUser"`
	Role         RoleResponse `json:"role,omitempty"`
	Job          JobResponse  `json:"job,omitempty"`
	AccessToken  string       `json:"accessToken,omitempty"`
	RefreshToken string       `json:"refreshToken,omitempty"`
}

type ForgotPassword struct {
	Email string `json:"email,omitempty" binding:"required"`
}
