package handlers

import (
	"encoding/json"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func RegisterCustomer(ctx *gin.Context, api types.APIFunc) {
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
	defer ctx.Request.Body.Close()
	_, err := api(data)
	if err != nil {
		responseErrors(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}
