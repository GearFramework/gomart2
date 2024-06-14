package handlers

import (
	"errors"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

func responseErrors(ctx *gin.Context, err error) {
	if errors.Is(err, types.ErrRegisterParamsRequest) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	} else if errors.Is(err, types.ErrInvalidAuthorization) ||
		errors.Is(err, types.ErrCustomerNotFound) ||
		errors.Is(err, types.ErrNeedAuthorization) ||
		errors.Is(err, types.ErrCustomerLogin) ||
		errors.Is(err, types.ErrCustomerNotFound) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if errors.Is(err, types.ErrInvalidOrderNumber) {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	} else if errors.Is(err, types.ErrOrderAlreadyExists) {
		ctx.AbortWithStatus(http.StatusOK)
		return
	} else if errors.Is(err, types.ErrOrderAnotherCustomer) ||
		errors.Is(err, types.ErrCustomerAlreadyExists) {
		ctx.AbortWithStatus(http.StatusConflict)
		return
	} else if errors.Is(err, types.ErrNotEnoughPoints) {
		ctx.AbortWithStatus(http.StatusPaymentRequired)
		return
	} else if errors.Is(err, types.ErrOrderAlreadyExists) ||
		errors.Is(err, types.ErrInvalidOrderNumber) {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
	}
	ctx.AbortWithStatus(http.StatusInternalServerError)
}
