package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room"
)

func main() {
	router := gin.Default()
	room.RegisterRouter(router.Group("/room"))
}
