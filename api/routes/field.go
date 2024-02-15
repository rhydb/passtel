package routes

import (
	"context"
	"net/http"
	"rhydb/passtel/api/schema"
	"rhydb/passtel/api/utils"

	"github.com/labstack/echo/v4"
)

type FieldParams struct {
	Type string `json:"type" validate:"required,oneof=username password otp url"`
        Value string `json:"value" validate:"required,max=127"`
}

func AddField(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(schema.User)

                item, err := getVaultItem(ctx, queries, user.UserID, c.Param("itemId"))
                if err != nil {
                        return err
                }

		params := new(FieldParams)
                if err = utils.BindValidateParams(c, params); err != nil {
                        return err
                }

                field, err := queries.AddField(ctx, schema.AddFieldParams{
                        ItemID: item.ItemID,
                        Type: schema.Fieldtype(params.Type),
                        Value: params.Value,
		})
                if err != nil {
			return echo.ErrInternalServerError
		}

		return c.JSON(http.StatusOK, field.FieldID)
	}
}

func ListItemFields(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                user := c.Get("user").(schema.User)
                vaultItem, err := getVaultItem(ctx, queries, user.UserID, c.Param("itemId"))
                if err != nil {
                        return err
                }

                fields, err := queries.GetItemFields(ctx, vaultItem.ItemID)
                if err != nil {
                        return utils.HandleQueryError(err, "No item or fields")
                }

                return c.JSON(http.StatusOK, fields)

        }
}

func UpdateField(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                fieldId, err := utils.GetIDParam(c.Param("fieldId"))
                if err != nil {
                        return err
                }

                userId, err := queries.GetFieldOwner(ctx, fieldId)
                if err != nil {
                        return echo.NewHTTPError(http.StatusNotFound, "No field that you own with that ID")
                }

                user := c.Get("user").(schema.User)
                if user.UserID != userId {
                        return echo.ErrUnauthorized
                }

                fieldUpdate := new(FieldParams)
                if err = utils.BindValidateParams(c, fieldUpdate); err != nil {
                        return err
                }

                err = queries.UpdateField(ctx, schema.UpdateFieldParams{
                        FieldID: fieldId,
                        Type: schema.Fieldtype(fieldUpdate.Type),
                        Value: fieldUpdate.Value,
                })
                if err != nil {
                        return utils.HandleQueryError(err, "Could not update field")
                }

                return c.NoContent(http.StatusOK)
        }
}

func DeleteField(ctx context.Context, queries *schema.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
                fieldId, err := utils.GetIDParam(c.Param("fieldId"))
                if err != nil {
                        return err
                }

                userId, err := queries.GetFieldOwner(ctx, fieldId)
                if err != nil {
                        return echo.NewHTTPError(http.StatusNotFound, "No field that you own with that ID")
                }

                user := c.Get("user").(schema.User)
                if user.UserID != userId {
                        return echo.ErrUnauthorized
                }

                if err = queries.DeleteField(ctx, fieldId); err != nil {
                        return utils.HandleQueryError(err, "No field with that ID")
                }
                
                return c.NoContent(http.StatusOK)
        }
}
