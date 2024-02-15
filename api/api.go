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

	db, err := sql.Open("postgres", "user=passtel password=passtel dbname=passtel sslmode=disable")
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
	vault.POST("/:vaultName", routes.CreateVault(ctx, queries))
	vault.GET("/:vaultId", routes.ListVaultItems(ctx, queries))
	vault.PUT("/:vaultId", routes.UpdateVault(ctx, queries))
	vault.DELETE("/:vaultId", routes.DeleteVault(ctx, queries))

	// vault items
        item := e.Group("/item", auth.TokenAuth(ctx, queries))
        item.POST("/:vaultId", routes.AddVaultItem(ctx, queries))
        item.PUT("/:itemId", routes.UpdateVaultItem(ctx, queries))
        item.DELETE("/:itemId", routes.DeleteVautlItem(ctx, queries))
        item.GET("/:itemId", routes.ListItemFields(ctx, queries))

        // fields
        field := e.Group("/field", auth.TokenAuth(ctx, queries))
        field.POST("/:itemId", routes.AddField(ctx, queries));
        field.PUT("/:fieldId", routes.UpdateField(ctx, queries))
        field.DELETE("/:fieldId", routes.DeleteField(ctx, queries))

	e.Logger.Fatal(e.Start(":1234"))
}
