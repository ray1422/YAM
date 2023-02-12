package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room"
	"github.com/ray1422/YAM-api/utils"
)

func main() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(utils.GetEnv("ALLOW_ORIGINS", "http://localhost:8080;http://localhost:3000"), ";"),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTION"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	apiRouter := router.Group("/api")
	room.RegisterRouter(apiRouter.Group("/room"))
	router.Run()
}
