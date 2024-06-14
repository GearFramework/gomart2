package handlers

import (
	"encoding/json"
	"errors"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func RegisterCustomer(ctx *gin.Context, api types.APIFunc) {
	defer ctx.Request.Body.Close()
	if !strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
		ctx.Status(http.StatusBadRequest)
		return
	}
	data := types.CustomerRegisterRequest{
		APIRequest: types.NewRequest(ctx),
	}
	if err := json.NewDecoder(ctx.Request.Body).Decode(&data); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err := api(data)
	if err != nil {
		if errors.Is(err, types.ErrCustomerAlreadyExists) {
			ctx.AbortWithStatus(http.StatusConflict)
			return
		} else if errors.Is(err, types.ErrRegisterParamsRequest) {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
	ctx.Status(http.StatusOK)
}
