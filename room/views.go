package room

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/utils/jwt"
	"github.com/rs/xid"
)

type roomIDPost struct {
	Password string `json:"password,omitempty"`
}

type view struct {
	hubModel Hub
}

type View interface {
	Views(roomGroup *gin.RouterGroup, baseURL string)
}

func (r view) Views(roomGroup *gin.RouterGroup, baseURL string) {
	r.hubModel.handleWS(roomGroup, "/")
	roomGroup.GET(baseURL, func(c *gin.Context) {
		// TODO authorization
		c.JSON(http.StatusOK, r.hubModel.roomList())
		// TODO Pagination
	})
	roomGroup.POST(baseURL, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"room_id": xid.New().String()})
	})
	roomGroup.GET(baseURL+":room_id/", func(c *gin.Context) {
		roomID, _ := c.Params.Get("room_id")
		_ = roomID
		// TODO verify roomID
		info, err := r.hubModel.roomInfo(roomID)
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
		body := roomIDPost{}
		c.BindJSON(&body)
		// TODO AUTH
		// TODO impl jwt pair (token & refresh)

		r.hubModel.roomCreate(roomID)

		token := jwt.New(48 * time.Hour)
		token.Payload["room_id"] = roomID
		c.JSON(http.StatusOK, map[string]interface{}{"token": token.TokenString(), "refresh": "EXAMPLE_REFRESH"})
	})

	roomGroup.DELETE(baseURL+":room_id/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"error": "NOT_IMPLEMENTED_YET"})
	})
}
