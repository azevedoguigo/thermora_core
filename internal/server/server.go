package server

import (
	"github.com/azevedoguigo/thermora_core/internal/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Config *config.Config
	Router *gin.Engine
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Config: cfg,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	r := gin.Default()

	r.GET("/health-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Service is running",
		})
	})

	return r
}

func (s *Server) Run() {
	s.Router = s.SetupRoutes()
	s.Router.Run(":" + s.Config.ServerPort)
}
