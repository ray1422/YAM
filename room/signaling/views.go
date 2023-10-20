package signaling

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	hubs           = map[string]*Hub{}
	globalHubsLock sync.RWMutex
)

// RoomWSHandler RoomWSHandler
func RoomWSHandler(router *gin.RouterGroup, baseURL string) {
	router.GET(baseURL+":room_id/ws/", func(c *gin.Context) {
		roomID := c.Param("room_id")
		globalHubsLock.Lock()
		defer globalHubsLock.Unlock()
		if hubs[roomID] == nil {
			if roomID == "neo" && os.Getenv("DEBUG") != "" { // TODO temp hardcoded
				hubs[roomID] = CreateHub(roomID)
			} else {
				c.JSON(http.StatusUnauthorized, nil)
				return
			}
		}
		hub := hubs[roomID]

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
