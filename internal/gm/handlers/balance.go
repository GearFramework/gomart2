package handlers

import (
	"errors"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetBalance(ctx *gin.Context, api types.APIFunc) {
	data := types.NewRequest(ctx)
	resp, err := api(data)
	if err != nil {
		if errors.Is(err, types.ErrInvalidAuthorization) ||
			errors.Is(err, types.ErrCustomerNotFound) ||
			errors.Is(err, types.ErrNeedAuthorization) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
