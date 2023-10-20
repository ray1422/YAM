package room

import (
	"github.com/gin-gonic/gin"
)

// RegisterRouter RegisterRouter
func RegisterRouter(roomGroup *gin.RouterGroup, r Hub) {
	v := view{hubModel: r}
	v.Views(roomGroup, "/")
}
