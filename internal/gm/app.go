package gm

import (
	"github.com/GearFramework/gomart2/internal/gm/config"
	"github.com/GearFramework/gomart2/internal/pkg/accrual"
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"github.com/GearFramework/gomart2/internal/pkg/auth"
	"github.com/GearFramework/gomart2/internal/pkg/db"
	"github.com/GearFramework/gomart2/internal/server"
	"github.com/GearFramework/gomart2/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GopherMartApp struct {
	Config  *config.GomartConfig
	Storage *db.Storage
	Server  *server.HTTPServer
	Auth    *auth.Auth
	Accrual *accrual.AccrualClient
	logger  *zap.SugaredLogger
}

func NewGomartApp(gomartConfig *config.GomartConfig) *GopherMartApp {
	return &GopherMartApp{
		Config: gomartConfig,
		logger: alog.NewLogger("info"),
	}
}

func (gm *GopherMartApp) Init() error {
	// init db storage
	gm.Storage = db.NewStorage(gm.Config.DatabaseDSN)
	if err := gm.Storage.Init(); err != nil {
		return err
	}
	// init auth component
	gm.Auth = auth.NewAuth()
	// init accrual client
	gm.Accrual = accrual.NewClient(gm.Config.AccrualAddr)
	// init server
	gm.Server = server.NewServer(server.NewServerConfig(gm.Config.Addr))
	gm.Server.SetMiddleware(func() gin.HandlerFunc {
		return middleware.Logger()
	})
	gm.Server.SetMiddleware(func() gin.HandlerFunc {
		return middleware.Compress()
	})
	//gm.Server.SetMiddleware(func() gin.HandlerFunc {
	//	return middleware.Auth(gm.Auth)
	//})
	gm.initRoutes()
	return nil
}

func (gm *GopherMartApp) Run() error {
	if err := gm.Server.Up(); err != nil {
		return err
	}
	return nil
}

func (gm *GopherMartApp) Stop() {
	if err := gm.Storage.Ping(); err == nil {
		gm.Storage.Close()
	}
}
