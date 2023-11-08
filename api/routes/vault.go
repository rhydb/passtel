package routes

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	auth "rhydb/passtel/api/handlers"
	"rhydb/passtel/api/schema"
	"strconv"

	"github.com/labstack/echo/v4"
)

// get a vault if it belongs to user, or throw unauthorised
func getUserVault(ctx context.Context, queries *schema.Queries, vaultId int64, userId int64) (schema.Vault, error) {
	vault, err := queries.GetVault(ctx, vaultId)
	if err != nil {
		return schema.Vault{}, echo.ErrBadRequest
	}

	if vault.UserID != userId {
		return schema.Vault{}, echo.ErrUnauthorized
	}

	return vault, nil
}

// convert the vault ID to an int64 or throw a bad request
func getVaultIdParam(param string) (int64, error) {
	vaultIdStr := param
	if vaultIdStr == "" {
		return -1, echo.ErrBadRequest
	}

	id, err := strconv.ParseInt(vaultIdStr, 10, 64)
	if err != nil {
		return -1, echo.ErrBadRequest
	}

	return id, nil
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
			return auth.HandleQueryError(err, "Vault already exists")
		}

		return c.JSON(http.StatusOK, echo.Map{
			"id": vault.VaultID,
		})
	}
}

func ListVaults(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(schema.User)
		vaults, err := queries.ListVaults(ctx, user.UserID)
		if err != nil {
			return auth.HandleQueryError(err, "No vaults found")
		}

		return c.JSON(http.StatusOK, vaults)
	}
}

type VaultUpdate struct {
	Name string `json:"name" validate:"required"`
}

func UpdateVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := getVaultIdParam(c.Param("id"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		vaultUpdate := new(VaultUpdate)
		if err = c.Bind(vaultUpdate); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(vaultUpdate); err != nil {
			return echo.ErrBadRequest
		}

		vault, err = queries.SetVaultName(ctx, schema.SetVaultNameParams{
			VaultID: vault.VaultID,
			Name:    vaultUpdate.Name,
		})
		if err != nil {
			return auth.HandleQueryError(err, "Vault already exists")
		}

		return c.NoContent(http.StatusOK)
	}
}

func DeleteVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := getVaultIdParam(c.Param("id"))
		if err != nil {
			return err
		}

		deletedId, err := queries.DeleteVault(ctx, vaultId)
		if err != nil || deletedId != vaultId {
                        log.Println("failed to delete vault:", err)
			return echo.ErrBadRequest
		}

		return c.NoContent(http.StatusOK)
	}
}

// vault items

func GetVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := getVaultIdParam(c.Param("id"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		vaultItems, err := queries.GetVaultItems(ctx, vault.VaultID)
		if err != nil {
			return echo.ErrBadRequest
		}

		return c.JSON(http.StatusOK, vaultItems)
	}
}

type VaultItemParams struct {
	Name string `json:"name" validate:"required,max=20"`
	Icon string `json:"icon"`
}

func AddVaultItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := getVaultIdParam(c.Param("id"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		params := new(VaultItemParams)
		if err = c.Bind(params); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(params); err != nil {
			return echo.ErrBadRequest
		}

		if err = queries.AddVaultItem(ctx, schema.AddVaultItemParams{
			VaultID: vault.VaultID,
			Name:    params.Name,
			Icon:    sql.NullString{String: params.Icon, Valid: true},
		}); err != nil {
			return echo.ErrInternalServerError
		}

		return c.NoContent(http.StatusOK)
	}
}

func UpdateVaultItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                vaultId, err := getVaultIdParam(c.Param("id")) 
                if err != nil {
                    return err
                }

                user := c.Get("user").(schema.User)
                vault, err := getUserVault(ctx, queries, vaultId, user.UserID)

		params := new(VaultItemParams)
		if err = c.Bind(params); err != nil {
			return echo.ErrBadRequest
		}

		if err = c.Validate(params); err != nil {
			return echo.ErrBadRequest
		}

		err = queries.UpdateVaultItem(ctx, schema.UpdateVaultItemParams{
			VaultID: vault.VaultID,
			Name:    params.Name,
			Icon:    sql.NullString{String: params.Icon, Valid: true},
		})
		if err != nil {
			return auth.HandleQueryError(err, "Vault item already exists")
		}

		return c.NoContent(http.StatusOK)
	}
}

func DeleteVautlItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                vaultItemId, err := getVaultIdParam(c.Param("id")) 
                if err != nil {
                    return err
                }

                user := c.Get("user").(schema.User)

                // get the vault item and vault data
                vaultItem, err := queries.GetVaultItem(vaultItemId)
                if err != nil {
                    return echo.ErrBadRequest
                }

                // the vault the item belongs to is not the user's
                if vaultItem.UserID != user.UserID {
                    return echo.ErrUnauthorized
                }

                err = queries.DeleteVaultItem(vaultItem.VaultID)
                if err != nil {
                    log.Println("Failed to delete vault item:", err)
                    return echo.ErrInternalServerError
                }

                return c.JSON(http.StatusOK, vaultItem)
		// return c.NoContent(http.StatusOK)
	}
}
