package gm

import (
	"github.com/GearFramework/gomart2/internal/gm/handlers"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/gin-gonic/gin"
)

func (gm *GopherMartApp) initRoutes() {
	gm.Server.Router.POST("/api/user/register", func(ctx *gin.Context) {
		handlers.RegisterCustomer(ctx, func(r types.Requester) (types.Response, error) {
			return gm.RegisterCustomer(r.(types.CustomerRegisterRequest))
		})
	})
	gm.Server.Router.POST("/api/user/login", func(ctx *gin.Context) {
		handlers.LoginCustomer(ctx, func(r types.Requester) (types.Response, error) {
			return gm.LoginCustomer(r.(types.CustomerLoginRequest))
		})
	})
	gm.Server.Router.POST("/api/user/orders", func(ctx *gin.Context) {
		handlers.AddOrder(ctx, func(r types.Requester) (types.Response, error) {
			return gm.AddOrder(r.(types.AddOrderRequest))
		})
	})
	gm.Server.Router.GET("/api/user/orders", func(ctx *gin.Context) {
		handlers.ListOrders(ctx, func(r types.Requester) (types.Response, error) {
			return gm.ListOrders(r.(types.APIRequest))
		})
	})
	gm.Server.Router.GET("/api/user/balance", func(ctx *gin.Context) {
		handlers.GetBalance(ctx, func(r types.Requester) (types.Response, error) {
			return gm.GetBalance(r.(types.APIRequest))
		})
	})
	gm.Server.Router.POST("/api/user/balance/withdraw", func(ctx *gin.Context) {
		handlers.Withdraw(ctx, func(r types.Requester) (types.Response, error) {
			return gm.Withdraw(r.(types.CustomerWithdrawRequest))
		})
	})
	gm.Server.Router.GET("/api/user/withdrawals", func(ctx *gin.Context) {
		handlers.ListWithdrawals(ctx, func(r types.Requester) (types.Response, error) {
			return gm.ListWithdrawals(r.(types.APIRequest))
		})
	})
	gm.Server.Router.GET("/ping", func(ctx *gin.Context) {
		handlers.Ping(ctx, func() error {
			return gm.Storage.Ping()
		})
	})
	gm.Server.Router.NoRoute(func(ctx *gin.Context) {
		gm.logger.Errorf("not found route %s", ctx.Request.URL)
		handlers.NotFound(ctx)
	})
}
