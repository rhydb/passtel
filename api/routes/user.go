package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"rhydb/passtel/api/schema"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Credentials struct {
	Username string `json:"username" validate:"required,max=25"`
	Password string `json:"password" validate:"required"`
}

func GetAuthedUser(c echo.Context) (err error) {
	user := c.Get("user").(schema.User)
	return c.JSON(http.StatusOK, user)
}

func GetToken(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		creds := new(Credentials)
		if err = c.Bind(creds); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(creds); err != nil {
			return echo.ErrBadRequest
		}

		creds.Password, err = hash(creds.Password)
		if err != nil {
			log.Println("failed to hash password:", err.Error())
			return echo.ErrInternalServerError
		}

		user, err := queries.CheckCreds(ctx, schema.CheckCredsParams{
			Username: creds.Username,
			Password: creds.Password,
		})
		if err != nil {
			return echo.ErrUnauthorized
		}

		token, err := queries.GenToken(ctx, user.UserID)
		if err != nil {
			log.Println("failed to generate token:", err)
			return echo.ErrInternalServerError
		}
		return c.JSON(http.StatusOK, echo.Map{
			"token": token,
		})
	}
}

func CreateUser(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		creds := new(Credentials)
		if err = c.Bind(creds); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(creds); err != nil {
			return echo.ErrBadRequest
		}

		creds.Password, err = hash(creds.Password)
		if err != nil {

			return echo.ErrInternalServerError
		}

		response, err := queries.CreateUser(ctx, schema.CreateUserParams{
			Username: creds.Username,
			Password: creds.Password,
		})
		if err != nil {
			pqError, isPQError := err.(*pq.Error)
			if isPQError && pqError.Constraint != "" {
				return c.JSON(http.StatusBadRequest, echo.Map{"message": "Username already exists"})
			}

			log.Println("failed to create new user:", err)
			return echo.ErrInternalServerError
		}

		return c.JSON(http.StatusOK, response)
	}
}

func hash(password string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(password))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
