package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping(ctx *gin.Context, api func() error) {
	if err := api(); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Data(http.StatusOK, "Content-Type: text/html", []byte("Pong"))
}
