package room

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room/signaling"
	"github.com/ray1422/YAM-api/utils/jwt"
	"github.com/rs/xid"
)

type roomIDPOST struct {
	Password string `json:"password,omitempty"`
}

func roomViews(roomGroup *gin.RouterGroup, baseURL string) {
	roomGroup.GET(baseURL, func(c *gin.Context) {
		// TODO authorization
		c.JSON(http.StatusOK, hubList())
		// TODO Pagination
	})
	roomGroup.POST(baseURL, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"room_id": xid.New().String()})
	})
	roomGroup.GET(baseURL+":room_id/", func(c *gin.Context) {
		roomID, _ := c.Params.Get("room_id")
		_ = roomID
		// TODO verify roomID
		info, err := hubInfo(roomID)
		if err != nil {
			c.JSON(http.StatusNotFound, nil)
			return
		}
		c.JSON(http.StatusOK, info)
	})
	roomGroup.POST(baseURL+":room_id/", func(c *gin.Context) {
		roomID, _ := c.Params.Get("room_id")
		_ = roomID
		// TODO verify roomID
		body := roomIDPOST{}
		c.BindJSON(&body)
		// TODO AUTH
		// TODO impl jwt pair (token & refresh)
		signaling.GlobalHubsLock.Lock()
		if signaling.Hubs[roomID] == nil {
			signaling.Hubs[roomID] = signaling.CreateHub(roomID)
		}

		signaling.GlobalHubsLock.Unlock()
		token := jwt.New(48 * time.Hour)
		token.Payload["room_id"] = roomID
		c.JSON(http.StatusOK, map[string]interface{}{"token": token.TokenString(), "refresh": "EXAMPLE_REFRESH"})
	})

	roomGroup.DELETE(baseURL+":room_id/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"error": "NOT_IMPLEMENTED_YET"})
	})
}
