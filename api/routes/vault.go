package routes

import (
	"context"
	"github.com/lib/pq"
	"net/http"
	"rhydb/passtel/api/schema"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		if idStr == "" {
			return echo.ErrBadRequest
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return echo.ErrBadRequest
		}

		vault, err := queries.GetVault(ctx, id)
		if err != nil {
			return echo.ErrBadRequest
		}

		return c.JSON(http.StatusOK, vault)
	}
}

func CreateVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return echo.ErrBadRequest
		}

		user := c.Get("user").(schema.User)

		vault, err := queries.CreateVault(ctx, schema.CreateVaultParams{
			Name:   name,
			UserID: user.UserID,
		})
		if err != nil {
			pqError, isPQError := err.(*pq.Error)
			if !isPQError || pqError.Constraint == "" {
				return echo.ErrInternalServerError
			}

			return echo.NewHTTPError(http.StatusBadRequest, "Vault already exists")
		}

		return c.JSON(http.StatusOK, echo.Map{
			"id": vault.VaultID,
		})
	}
}
