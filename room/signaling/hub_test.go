package signaling

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientJoinAfterHubClose(t *testing.T) {
	oldCleanerTimeout := cleanerTimeout
	cleanerTimeout = 0 // mock
	t.Cleanup(func() {
		cleanerTimeout = oldCleanerTimeout
	})
	GlobalHubsLock.Lock()
	Hubs["www"] = CreateHub("www")
	hub := Hubs["www"]
	GlobalHubsLock.Unlock()
	c1 := hub.NewClient(nil)
	time.Sleep(300 * time.Millisecond) // make it timeout
	assert.NotNil(t, c1.registerClient([]byte(`{"token": "www"}`)))
	time.Sleep(300 * time.Millisecond) // make it timeout
}
