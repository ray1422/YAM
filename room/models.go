package room

import (
	"github.com/gin-gonic/gin"
	"github.com/ray1422/YAM-api/room/signaling"
)

type hub struct {
	sigSrv signaling.Server
}
type HubModel interface {
	roomList() []string
	roomInfo(hubID string) (*signaling.HubInfo, error)
	roomCreate(hubID string) *signaling.Hub
	handleWS(router *gin.RouterGroup, baseURL string)
}

func (r hub) roomList() []string {
	return r.sigSrv.RoomList()
}

func (r hub) roomInfo(hubID string) (*signaling.HubInfo, error) {
	return r.sigSrv.RoomInfoByID(hubID)
}

func (r hub) roomCreate(hubID string) *signaling.Hub {
	return r.sigSrv.RoomCreate(hubID)
}

func (r hub) handleWS(router *gin.RouterGroup, baseURL string) {
	r.sigSrv.RoomWSHandler(router, baseURL)
}

// NewHub creates a new hub model
func NewHub() HubModel {
	return hub{sigSrv: signaling.New()}

}
