package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GearFramework/gomart/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Withdraw(ctx *gin.Context, api types.APIFunc) {
	defer ctx.Request.Body.Close()
	if !strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
		fmt.Println("invalid request type")
		ctx.Status(http.StatusBadRequest)
		return
	}
	data := types.CustomerWithdrawRequest{
		APIRequest: types.NewRequest(ctx),
	}
	if err := json.NewDecoder(ctx.Request.Body).Decode(&data); err != nil {
		fmt.Println("error decoded json request;", err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err := api(data)
	if err != nil {
		if errors.Is(err, types.ErrInvalidAuthorization) ||
			errors.Is(err, types.ErrCustomerNotFound) ||
			errors.Is(err, types.ErrNeedAuthorization) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if errors.Is(err, types.ErrNotEnoughPoints) {
			ctx.AbortWithStatus(http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, types.ErrOrderAlreadyExists) || errors.Is(err, types.ErrInvalidOrderNumber) {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		}
		fmt.Println("internal error;", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}
