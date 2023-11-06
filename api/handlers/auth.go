package auth

import (
	"context"
	"log"
	"net/http"
	"rhydb/passtel/api/schema"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
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

func HandleQueryError(err error, msg string) error {
	pqError, isPQError := err.(*pq.Error)
	if !isPQError || pqError.Constraint == "" {
		log.Println("pq error:", pqError.Message)
		return echo.ErrInternalServerError
	}

	return echo.NewHTTPError(http.StatusBadRequest, msg)
}
