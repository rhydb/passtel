package utils

import (
	"strconv"
	"net/http"
	"github.com/lib/pq"
        "log"
	"github.com/labstack/echo/v4"
)

func HandleQueryError(err error, msg string) error {
	pqError, isPQError := err.(*pq.Error)
	if !isPQError || pqError.Constraint == "" {
		log.Println("pq error:", pqError.Message)
		return echo.ErrInternalServerError
	}

	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

// convert an parameter to an int64 that can be used as an ID or throw a bad request
func GetIDParam(param string) (int64, error) {
	idStr := param
	if idStr == "" {
                err := echo.ErrBadRequest
                err.Message = "Empty param: " + param
		return -1, err
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
                err := echo.ErrBadRequest
                err.Message = "Invalid param: " + param
                return -1, err
	}

	return id, nil
}

func BindValidateParams(c echo.Context, params interface{}) (err error) {
        if err = c.Bind(params); err != nil {
            return echo.ErrBadRequest
        }

        if err = c.Validate(params); err != nil {
            return echo.ErrBadRequest
        }

        return nil
}
