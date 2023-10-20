package room

import (
	"github.com/gin-gonic/gin"
)

// RegisterRouter RegisterRouter
func RegisterRouter(roomGroup *gin.RouterGroup, r HubModel) {
	v := view{hubModel: r}
	v.Views(roomGroup, "/")
}
