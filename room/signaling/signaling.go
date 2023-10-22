package signaling

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type server struct {
	rooms     map[string]*Room
	roomsLock sync.RWMutex
}

// Server is the signaling server holding all the rooms
type Server interface {
	RoomList() []string
	RoomInfoByID(roomID string) (*RoomInfo, error)
	// RoomCreate(roomID string) *Room
	RoomCreate(roomID string) *Room
	RoomWSHandler(router *gin.RouterGroup, baseURL string)
}

var _ Server = (*server)(nil)

// New creates a new signaling server
func New() Server {
	return &server{
		rooms: map[string]*Room{},
	}
}

func new() *server {
	return &server{
		rooms: map[string]*Room{},
	}
}
