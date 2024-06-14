package handlers

import (
	"errors"
	"github.com/GearFramework/gomart/internal/gm/types"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func AddOrder(ctx *gin.Context, api types.APIFunc) {
	defer ctx.Request.Body.Close()
	if !strings.Contains(ctx.Request.Header.Get("Content-Type"), "text/plain") {
		ctx.Status(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	data := types.AddOrderRequest{
		APIRequest:  types.NewRequest(ctx),
		OrderNumber: string(body),
	}
	_, err = api(data)
	if err != nil {
		if errors.Is(err, types.ErrInvalidOrderNumber) {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		} else {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
	ctx.Status(http.StatusOK)
}
