package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"sinarlog.com/internal/app/usecase"
)

// PaginateMiddleware verifies whether Order field in a pagination
// request matches a given set of column names. This is to prevent
// an error during query. It also  returns a ClientError code instead
// of InternalServerError.
func (m *Middleware) PaginateMiddleware(columns ...any) gin.HandlerFunc {
	initCol := []any{"id", "created_at", "updated_at"}

	return func(c *gin.Context) {
		cols := append(initCol, columns...)
		preq := m.ParsePagination(c)
		if err := validation.Validate(preq.Order,
			validation.When(preq.Order != "created_at",
				validation.In(cols...).Error(
					fmt.Sprintf("unknown column name detected in order by option: %s", preq.Order))),
		); err != nil {
			m.ClientError(c, usecase.NewClientError("Pagination Params", err))
			c.Abort()
		}

		m.addToContext(c, "pagination", preq)
		c.Next()
	}
}
