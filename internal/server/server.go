package server

import (
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type HTTPServer struct {
	HTTP   *http.Server
	Router *gin.Engine
	Logger *zap.SugaredLogger
	Config *Config
}

type MiddlewareFunc func() gin.HandlerFunc

func NewServer(conf *Config) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	return &HTTPServer{
		Config: conf,
		Logger: alog.NewLogger("Server " + conf.Addr),
		Router: gin.New(),
	}
}

func (serv *HTTPServer) SetMiddleware(mw MiddlewareFunc) *HTTPServer {
	serv.Router.Use(mw())
	return serv
}

func (serv *HTTPServer) Init(initRoutes func()) error {
	initRoutes()
	return nil
}

func (serv *HTTPServer) Up() error {
	serv.HTTP = &http.Server{
		Addr:    serv.Config.Addr,
		Handler: serv.Router,
	}
	serv.Logger.Infof("start server at the: %s", serv.Config.Addr)
	err := serv.HTTP.ListenAndServe()
	if err != nil {
		serv.Logger.Errorf("failed: %s", err.Error())
		return err
	}
	return nil
}
