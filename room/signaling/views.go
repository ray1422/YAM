package signaling

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	hubs = map[string]*Hub{}
)

func RoomWS(router *gin.RouterGroup, baseURL string) {
	router.GET(baseURL+":room_id/ws/", func(c *gin.Context) {
		if hubs[c.Param("room_id")] == nil {
			hubs[c.Param("room_id")] = CreateHub()
		}
		hub := hubs[c.Param("room_id")]
		upgrader := websocket.Upgrader{
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
		}
		hub.NewClient(conn)
	})
}
