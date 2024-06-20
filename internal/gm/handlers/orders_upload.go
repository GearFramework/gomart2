package handlers

import (
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func AddOrder(ctx *gin.Context, api types.APIFunc) {
	if !strings.Contains(ctx.Request.Header.Get("Content-Type"), "text/plain") {
		ctx.Status(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer ctx.Request.Body.Close()
	data := types.AddOrderRequest{
		APIRequest:  types.NewRequest(ctx),
		OrderNumber: string(body),
	}
	_, err = api(data)
	if err != nil {
		responseErrors(ctx, err)
		return
	}
	ctx.Status(http.StatusAccepted)
}
