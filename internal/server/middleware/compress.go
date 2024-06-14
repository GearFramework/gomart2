package middleware

import (
	"github.com/GearFramework/gomart2/internal/pkg/compresser"
	"github.com/gin-gonic/gin"
)

func Compress() gin.HandlerFunc {
	return compresser.NewCompressor()
}
