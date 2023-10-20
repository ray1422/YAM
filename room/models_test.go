package room

import (
	"testing"

	"github.com/ray1422/YAM-api/room/signaling"
	"github.com/stretchr/testify/assert"
)

func TestListRoom(t *testing.T) {
	// create hubs
	signaling.CreateHub("www")
	signaling.CreateHub("asdf")

	ss := hubList()
	assert.ElementsMatch(t, []string{"www", "asdf"}, ss)
}
