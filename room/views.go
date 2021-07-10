package room

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

type RoomIDPOST struct {
	password string
}

func roomViews(roomGroup *gin.RouterGroup, baseURL string) {
	roomGroup.GET(baseURL, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"error": "NOT_IMPLEMENTED_YET"})
		// TODO Pagination
	})
	roomGroup.POST(baseURL, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"room_id": xid.New().String()})
	})
	roomGroup.GET(baseURL+":room_id/", func(c *gin.Context) {
		roomID, _ := c.Params.Get("room_id")
		_ = roomID
		// TODO verify roomID
		c.JSON(http.StatusOK, map[string]interface{}{"members": []string{}, "error": "NOT_IMPLEMENTED_YET"})
	})
	roomGroup.POST(baseURL+":room_id/", func(c *gin.Context) {
		roomID, _ := c.Params.Get("room_id")
		_ = roomID
		// TODO verify roomID
		body := RoomIDPOST{}
		c.BindJSON(&body)
		// TODO AUTH
		// TODO impelement JWT token
		c.JSON(http.StatusOK, map[string]interface{}{"token": "EXAMPLE_TOKEN", "refresh": "EXAMPLE_REFRESH"})
	})

	roomGroup.DELETE(baseURL+":room_id/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"members": []string{}, "error": "NOT_IMPLEMENTED_YET"})
	})
}
