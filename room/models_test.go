package room

import (
	"testing"

	"github.com/ray1422/YAM-api/room/signaling"
	"github.com/stretchr/testify/assert"
)

func TestListRoom(t *testing.T) {
	room := hub{sigSrv: signaling.New()}
	// create hubs
	room.sigSrv.RoomCreate("www")
	room.sigSrv.RoomCreate("asdf")

	ss := room.roomList()
	assert.ElementsMatch(t, []string{"www", "asdf"}, ss)
}
