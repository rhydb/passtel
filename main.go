package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"rhydb/passtel/schema"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

func generateToken(userId int64) (uuid.UUID, error) {
	token, err := queries.GenToken(ctx, userId)
	if err != nil {
		return uuid.UUID{}, err
	}

	return token, nil
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	db, err := sql.Open("postgres", "user=rb dbname=schema sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			data := new(struct {
				Token string `json:"token"`
			})
			if err = c.Bind(data); err != nil {
				//c.JSON(http.StatusBadRequest, echo.Map{"message": "improper token"})
			}

			//token, err := uuid.Parse(data.Token)
			//if err != nil {
			//	log.Println(err.Error())
			//	return c.JSON(http.StatusBadRequest, echo.Map{"message": "improper token"})
			//}

			//user, err := queries.CheckToken(ctx, token)
			//if err != nil {
			//	log.Println(err.Error())
			//	return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid token"})
			//}

			//c.Set("user", user)
			return next(c)
		}
	}

	queries = schema.New(db)

	e.POST("/register", func(c echo.Context) (err error) {
		data := new(User)
		if err = c.Bind(data); err != nil {
			return errorJson(c, err)
		}

		if err = c.Validate(data); err != nil {
			return err
		}

		data.Password, err = hash(data.Password)
		if err != nil {
			return errorJson(c, err)
		}

		response, err := queries.CreateUser(ctx, schema.CreateUserParams{
			Username: data.Username,
			Password: data.Password,
		})
		if err != nil {
			return errorJson(c, err)
		}

		return c.JSON(http.StatusOK, response)
	})

	e.POST("/login", func(c echo.Context) error {
		data := new(User)
		if err = c.Bind(data); err != nil {
			return errorJson(c, err)
		}

		if err = c.Validate(data); err != nil {
			return err
		}

		data.Password, err = hash(data.Password)
		if err != nil {
			return errorJson(c, err)
		}

		user, err := queries.CheckCreds(ctx, schema.CheckCredsParams{
			Username: data.Username,
			Password: data.Password,
		})
		if err != nil {
			return errorJson(c, err)
		}

		token, err := generateToken(user.ID)
		if err != nil {
			return errorJson(c, err)
		}
		return c.JSON(http.StatusOK, echo.Map{
			"token": token,
		})
	})

	e.GET("/user/:id", func(c echo.Context) (err error) {
		params := struct {
			Id int64 `json:"id" param:"id"`
		}{}

		if err = c.Bind(&params); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
		}

		user, err := queries.GetUser(ctx, params.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, user)

	}, authMiddleware)

	e.GET("/users", func(c echo.Context) error {
		// list all users
		users, err := queries.ListUsers(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		return c.JSON(http.StatusOK, users)
	})

	e.Logger.Fatal(e.Start(":1234"))
}
