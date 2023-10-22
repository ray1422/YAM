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
		s.roomsLock.Lock()
		defer s.roomsLock.Unlock()
		if s.rooms[roomID] == nil {
			if roomID == "neo" && os.Getenv("DEBUG") != "" { // TODO temp hardcoded
				s.roomsLock.Unlock()
				s.rooms[roomID] = s.RoomCreate(roomID)
				s.roomsLock.Lock()
			} else {
				c.JSON(http.StatusUnauthorized, nil)
				return
			}
		}
		room := s.rooms[roomID]

		upgrader := websocket.Upgrader{
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
		}
		room.NewClient(conn)
	})

}
