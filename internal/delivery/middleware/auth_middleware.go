package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"sinarlog.com/internal/app/usecase"
)

type AuthHeader struct {
	Authorization string `header:"Authorization" binding:"required"`
}

func (m *Middleware) AuthMiddleware(uc usecase.ICredentialUseCase, roles ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var auth AuthHeader
		if err := c.Copy().ShouldBindHeader(&auth); err != nil {
			m.Unauthorized(c, usecase.NewUnauthorizedError(fmt.Errorf("this is a protected endpoint, it requires an auth token")))
			return
		}

		auth.Authorization = strings.ReplaceAll(auth.Authorization, "Bearer ", "")
		user, err := uc.Authorize(c.Request.Context(), auth.Authorization, roles...)
		if err != nil {
			m.Unauthorized(c, err)
			return
		}

		m.addToContext(c, "user", user)
	}
}
