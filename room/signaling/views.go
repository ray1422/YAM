package signaling

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// RoomWSHandler RoomWSHandler
func (s *server) RoomWSHandler(router *gin.RouterGroup, baseURL string) {
	router.GET(baseURL+":room_id/ws/", func(c *gin.Context) {
		roomID := c.Param("room_id")
		s.hubLock.Lock()
		defer s.hubLock.Unlock()
		if s.hubs[roomID] == nil {
			if roomID == "neo" && os.Getenv("DEBUG") != "" { // TODO temp hardcoded
				s.hubLock.Unlock()
				s.hubs[roomID] = s.RoomCreate(roomID)
				s.hubLock.Lock()
			} else {
				c.JSON(http.StatusUnauthorized, nil)
				return
			}
		}
		hub := s.hubs[roomID]

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
