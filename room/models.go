package room

import (
	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room/signaling"
)

type hub struct {
	sigSrv signaling.Server
}

// Hub manages rooms
type Hub interface {
	roomList() []string
	roomInfo(hubID string) (*signaling.RoomInfo, error)
	roomCreate(hubID string) *signaling.Room
	handleWS(router *gin.RouterGroup, baseURL string)
}

func (r hub) roomList() []string {
	return r.sigSrv.RoomList()
}

func (r hub) roomInfo(hubID string) (*signaling.RoomInfo, error) {
	return r.sigSrv.RoomInfoByID(hubID)
}

func (r hub) roomCreate(hubID string) *signaling.Room {
	return r.sigSrv.RoomCreate(hubID)
}

func (r hub) handleWS(router *gin.RouterGroup, baseURL string) {
	r.sigSrv.RoomWSHandler(router, baseURL)
}

// NewHub creates a new hub model
func NewHub() Hub {
	return hub{sigSrv: signaling.New()}

}
