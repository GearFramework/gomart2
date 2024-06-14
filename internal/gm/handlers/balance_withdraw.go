package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/GearFramework/gomart2/internal/gm/types"
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
		responseErrors(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}
