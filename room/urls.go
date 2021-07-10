package room

import "github.com/gin-gonic/gin"

func RegisterRouter(roomGroup *gin.RouterGroup) {
	roomViews(roomGroup, "/")
}
