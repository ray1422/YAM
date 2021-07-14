package signaling

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientJoinAfterHubClose(t *testing.T) {
	cleanerTimeout = 0 // mock
	globalHubsLock.Lock()
	hubs["www"] = CreateHub("www")
	hub := hubs["www"]
	globalHubsLock.Unlock()
	c1 := hub.NewClient(nil)
	time.Sleep(300 * time.Millisecond) // make it timeout
	assert.NotNil(t, c1.registerClient([]byte(`{"token": "www"}`)))
	time.Sleep(300 * time.Millisecond) // make it timeout
}
