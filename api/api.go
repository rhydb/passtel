package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"rhydb/passtel/api/handlers"
	"rhydb/passtel/api/routes"
	"rhydb/passtel/api/schema"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var queries *schema.Queries
var ctx = context.Background()

type (
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
	defer db.Close()

	queries = schema.New(db)

	e.POST("/register", routes.CreateUser(ctx, queries))
	e.POST("/login", routes.GetToken(ctx, queries))
	e.GET("/user", routes.GetAuthedUser, auth.TokenAuth(ctx, queries))

	// all vault routes require authentication
	e.GET("/vaults", routes.ListVaults(ctx, queries), auth.TokenAuth(ctx, queries))

	vault := e.Group("/vault", auth.TokenAuth(ctx, queries))
	vault.POST("/:name", routes.CreateVault(ctx, queries))
	vault.GET("/:id", routes.GetVault(ctx, queries))
	vault.PUT("/:id", routes.UpdateVault(ctx, queries))
	vault.DELETE("/:id", routes.DeleteVault(ctx, queries))

	// vault items
	vault.POST("/:id", routes.AddVaultItem(ctx, queries))
        vault.PUT(":/id", routes.UpdateVaultItem(ctx, queries))
        vault.DELETE("/:id")

	e.Logger.Fatal(e.Start(":1234"))
}
