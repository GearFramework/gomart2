package handlers

import (
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ListOrders(ctx *gin.Context, api types.APIFunc) {
	data := types.NewRequest(ctx)
	resp, err := api(data)
	if err != nil {
		responseErrors(ctx, err)
		return
	}
	if len(resp.([]types.Order)) == 0 {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
