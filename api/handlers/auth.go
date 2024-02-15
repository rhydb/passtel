package auth

import (
	"context"
	"rhydb/passtel/api/schema"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TokenAuth(ctx context.Context, queries *schema.Queries) echo.MiddlewareFunc {
	return middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		token, err := uuid.Parse(key)
		if err != nil {
			return false, err
		}

		row, err := queries.CheckToken(ctx, token)
		if err != nil {
			return false, err
		}

		c.Set("user", schema.User{
			UserID:   row.UserID,
			Username: row.Username,
			Password: row.Password,
		})
		return true, nil
	})
}
