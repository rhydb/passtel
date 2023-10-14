package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"rhydb/passtel/schema"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var queries *schema.Queries
var ctx = context.Background()

func hash(password string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(password))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type (
	User struct {
		Username string `json:"username" validate:"required,max=25"`
		Password string `json:"password" validate:"required"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func errorJson(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
}

type AuthData struct {
	Token string `json:"token"`
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	db, err := sql.Open("postgres", "user=rb dbname=passtel sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	authMiddleware := middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
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

	queries = schema.New(db)

	e.POST("/register", func(c echo.Context) (err error) {
		data := new(User)
		if err = c.Bind(data); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(data); err != nil {
			return echo.ErrBadRequest
		}

		data.Password, err = hash(data.Password)
		if err != nil {
			log.Println("failed to hash password:", err.Error())
			return echo.ErrInternalServerError
		}

		response, err := queries.CreateUser(ctx, schema.CreateUserParams{
			Username: data.Username,
			Password: data.Password,
		})
		if err != nil {
			log.Println("failed to create new user:", err)
			return echo.ErrInternalServerError
		}

		return c.JSON(http.StatusOK, response)
	})

	e.POST("/login", func(c echo.Context) error {
		data := new(User)
		if err = c.Bind(data); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(data); err != nil {
			return echo.ErrBadRequest
		}

		data.Password, err = hash(data.Password)
		if err != nil {
			log.Println("failed to hash password:", err.Error())
			return echo.ErrInternalServerError
		}

		user, err := queries.CheckCreds(ctx, schema.CheckCredsParams{
			Username: data.Username,
			Password: data.Password,
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
	})

	e.GET("/user", func(c echo.Context) (err error) {
		user := c.Get("user").(schema.User)
		log.Println("user=", user)
		return c.JSON(http.StatusOK, user)
	}, authMiddleware)

	e.GET("/users", func(c echo.Context) error {
		// list all users
		users, err := queries.ListUsers(ctx)
		if err != nil {
			log.Println(err)
			return echo.ErrBadRequest
		}
		return c.JSON(http.StatusOK, users)
	})

	vault := e.Group("/vault", authMiddleware)
	vault.GET("/:id", func(c echo.Context) error {
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
	})

	// create a new vault
	vault.POST("/:name", func(c echo.Context) error {
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
	})

	e.Logger.Fatal(e.Start(":1234"))
}
