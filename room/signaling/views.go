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
	// TODO use sync.Map
	Hubs           = map[string]*Hub{}
	GlobalHubsLock sync.RWMutex
)

// RoomWS RoomWS
func RoomWS(router *gin.RouterGroup, baseURL string) {
	router.GET(baseURL+":room_id/ws/", func(c *gin.Context) {
		roomID := c.Param("room_id")
		GlobalHubsLock.Lock()
		if Hubs[roomID] == nil {
			if roomID == "neo" && os.Getenv("DEBUG") != "" { // TODO temp hardcoded
				Hubs[roomID] = CreateHub(roomID)
			} else {
				c.JSON(http.StatusUnauthorized, nil)
				return
			}
		}
		hub := Hubs[roomID]
		GlobalHubsLock.Unlock()
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
