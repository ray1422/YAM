package signaling

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type server struct {
	hubs    map[string]*Hub
	hubLock sync.RWMutex
}

// Server is the signaling server holding all the hubs
type Server interface {
	RoomList() []string
	RoomInfoByID(hubID string) (*HubInfo, error)
	// RoomCreate(roomID string) *Hub
	RoomCreate(roomID string) *Hub
	RoomWSHandler(router *gin.RouterGroup, baseURL string)
}

var _ Server = (*server)(nil)

// New creates a new signaling server
func New() Server {
	return &server{
		hubs: map[string]*Hub{},
	}
}

func new() *server {
	return &server{
		hubs: map[string]*Hub{},
	}
}
