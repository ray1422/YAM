package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room"
)

func main() {
	router := gin.Default()
	apiRouter := router.Group("/api")
	room.RegisterRouter(apiRouter.Group("/room"))
	router.Run()
}
