package signaling

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	hubs           = map[string]*Hub{}
	globalHubsLock sync.RWMutex
)

func RoomWS(router *gin.RouterGroup, baseURL string) {
	router.GET(baseURL+":room_id/ws/", func(c *gin.Context) {
		roomID := c.Param("room_id")
		globalHubsLock.Lock()
		if hubs[roomID] == nil {
			hubs[roomID] = CreateHub(roomID)
		}
		hub := hubs[roomID]
		globalHubsLock.Unlock()
		upgrader := websocket.Upgrader{
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
		}
		hub.NewClient(conn)
	})
}
