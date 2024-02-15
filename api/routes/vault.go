package routes

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"rhydb/passtel/api/schema"
	"rhydb/passtel/api/utils"

	"github.com/labstack/echo/v4"
)

// get a vault if it belongs to user, or throw unauthorised
func getUserVault(ctx context.Context, queries *schema.Queries, vaultId int64, userId int64) (schema.Vault, error) {
	vault, err := queries.GetVault(ctx, vaultId)
        if err == sql.ErrNoRows {
            return schema.Vault{}, echo.NewHTTPError(http.StatusNotFound, "No vault with that ID")
        }

	if err != nil {
                log.Println("Error finding vault:", err)
		return schema.Vault{}, echo.ErrBadRequest
	}

	if vault.UserID != userId {
		return schema.Vault{}, echo.ErrUnauthorized
	}

	return vault, nil
}


func getVaultItem(ctx context.Context, queries *schema.Queries, userId int64, itemIdParam string) (schema.VaultItem, error) {

        // get the item ID from the route
        itemId, err := utils.GetIDParam(itemIdParam) 
        if err != nil {
            return schema.VaultItem{}, err
        }

        // get the vault item
        vaultItem, err := queries.GetVaultItem(ctx, itemId)
        if err == sql.ErrNoRows {
            return schema.VaultItem{}, echo.NewHTTPError(http.StatusNotFound, "No item with that ID")
        }

        // get the users vault using the item
        vault, err := getUserVault(ctx, queries, vaultItem.VaultID, userId)
        if err != nil {
            return schema.VaultItem{}, err
        }

        if vault.UserID != userId {
                return schema.VaultItem{}, echo.ErrUnauthorized
        }

        if err != nil || vaultItem.VaultID != vault.VaultID {
            return schema.VaultItem{}, echo.ErrBadRequest
        }

        return schema.VaultItem{
            ItemID: vaultItem.ItemID,
            VaultID: vaultItem.VaultID,
            Name: vaultItem.Name,
            Icon: vaultItem.Icon,
        }, nil
}

// create an empty vault with a name
func CreateVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("vaultName")
		if name == "" {
			return echo.ErrBadRequest
		}

		user := c.Get("user").(schema.User)

		vault, err := queries.CreateVault(ctx, schema.CreateVaultParams{
			Name:   name,
			UserID: user.UserID,
		})
		if err != nil {
			return utils.HandleQueryError(err, "Vault already exists")
		}

		return c.JSON(http.StatusOK, echo.Map{
			"vaultId": vault.VaultID,
		})
	}
}

// list a user's vaults
func ListVaults(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(schema.User)
		vaults, err := queries.ListVaults(ctx, user.UserID)
		if err != nil {
			return utils.HandleQueryError(err, "No vaults found")
		}

		return c.JSON(http.StatusOK, vaults)
	}
}


// update a user's vault name
func UpdateVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
        type VaultUpdate struct {
            Name string `json:"name" validate:"required"`
        }

	return func(c echo.Context) error {
		vaultId, err := utils.GetIDParam(c.Param("vaultId"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		vaultUpdate := new(VaultUpdate)
                if err = utils.BindValidateParams(c, vaultUpdate); err != nil {
                        return err;
                }

		vault, err = queries.SetVaultName(ctx, schema.SetVaultNameParams{
			VaultID: vault.VaultID,
			Name:    vaultUpdate.Name,
		})
		if err != nil {
			return utils.HandleQueryError(err, "Vault already exists")
		}

		return c.NoContent(http.StatusOK)
	}
}

// delete a user's vault and everything in it
func DeleteVault(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := utils.GetIDParam(c.Param("vaultId"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)
		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		deletedId, err := queries.DeleteVault(ctx, vault.VaultID)
                if err == sql.ErrNoRows {
                    return echo.NewHTTPError(http.StatusNotFound, "No vault with that ID")
                }

		if err != nil || deletedId != vaultId {
                        return utils.HandleQueryError(err, "ID mismatch")
		}

		return c.NoContent(http.StatusOK)
	}
}

// vault items

// list the items in one vault
func ListVaultItems(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := utils.GetIDParam(c.Param("vaultId"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		vaultItems, err := queries.GetVaultItems(ctx, vault.VaultID)
                if err == sql.ErrNoRows {
                    return echo.ErrNotFound
                }
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

// create an item in a vault
func AddVaultItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		vaultId, err := utils.GetIDParam(c.Param("vaultId"))
		if err != nil {
			return err
		}

		user := c.Get("user").(schema.User)

		vault, err := getUserVault(ctx, queries, vaultId, user.UserID)
		if err != nil {
			return err
		}

		params := new(VaultItemParams)
		if err = utils.BindValidateParams(c, params); err != nil {
			return err
		}

                item, err := queries.AddVaultItem(ctx, schema.AddVaultItemParams{
			VaultID: vault.VaultID,
			Name:    params.Name,
			Icon:    &params.Icon,
		})
                if err != nil {
                        return utils.HandleQueryError(err, "Invalid vault")
		}

		return c.JSON(http.StatusOK, echo.Map{
                        "itemId": item.ItemID,
                })
	}
}

// update an item in a vault
func UpdateVaultItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                user := c.Get("user").(schema.User)
                vaultItem, err := getVaultItem(ctx, queries, user.UserID, c.Param("itemId"))
                if err != nil {
                    return err
                }

                // get the update info from the request
		params := new(VaultItemParams)
		if err = utils.BindValidateParams(c, params); err != nil {
			return err
		}

                // default the update info back to the original data if not passed
                if len(params.Name) == 0 {
                    params.Name = vaultItem.Name
                }
                if len(params.Icon) == 0 && vaultItem.Icon != nil {
                    params.Icon = *vaultItem.Icon
                }

		err = queries.UpdateVaultItem(ctx, schema.UpdateVaultItemParams{
			VaultID: vaultItem.VaultID,
                        ItemID: vaultItem.ItemID,
			Name:    params.Name,
			Icon:    &params.Icon,
		})
		if err != nil {
			return utils.HandleQueryError(err, "Vault item already exists")
		}

		return c.NoContent(http.StatusOK)
	}
}

// delete an item from a vault
func DeleteVautlItem(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                user := c.Get("user").(schema.User)
                vaultItem, err := getVaultItem(ctx, queries, user.UserID, c.Param("itemId"))
                if err != nil {
                    return err
                }

                err = queries.DeleteVaultItem(ctx, vaultItem.ItemID)
                if err != nil {
                        return utils.HandleQueryError(err, "No vault with that ID")
                }

                return c.JSON(http.StatusOK, vaultItem)
	}
}

