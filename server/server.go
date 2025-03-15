package server

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"lyonbot.github.com/my_app/metrics"
	"lyonbot.github.com/my_app/misc"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(listenOn string) error {
	r := gin.Default()
	r.SetTrustedProxies(strings.Fields(os.Getenv("TRUSTED_PROXIES")))
	r.Use(CORSMiddleware())
	r.Use(IPRateLimitMiddleware())
	r.Use(metrics.PrometheusMiddleware())

	setupStaticAssets(r)

	r.GET(misc.Getenv("METRICS_PATH", "/metrics"), gin.WrapH(metrics.GetHttpHandler()))

	r.GET("/api/hello", s.handleHello)
	// r.POST("/api/deleteHistory", JWTMiddleware(), s.handleDeleteHistory)

	return r.Run(misc.DefaultString(listenOn, ":8080"))
}

func (s *Server) handleHello(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello there",
		"ip":      c.ClientIP(),
		"headers": c.Request.Header,
	})
}
