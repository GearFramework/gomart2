package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NotFound(ctx *gin.Context) {
	ctx.Status(http.StatusBadRequest)
}
