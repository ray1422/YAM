package signaling

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestClientJoinAfterHubClose(t *testing.T) {
	oldCleanerTimeout := cleanerTimeout
	cleanerTimeout = 100 * time.Millisecond // mock
	t.Cleanup(func() {
		cleanerTimeout = oldCleanerTimeout
	})
	hubs["www"] = CreateHub("www")
	hub := hubs["www"]
	c1 := hub.NewClient(nil)
	time.Sleep(300 * time.Millisecond) // make it timeout
	assert.NotNil(t, c1.registerClient([]byte(`{"token": "www"}`)))
	time.Sleep(300 * time.Millisecond) // make it timeout
}

func TestHubInfo(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()
	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()
	CreateHub("www")
	CreateHub("asdf")
	h := hubs["www"]
	c := h.NewClient(ws)
	h.RegisterChan <- c
	hi, err := HubInfoByID("www")
	assert.Nil(t, err)
	assert.Equal(t, hi.ID, "www")
	assert.Equal(t, 1, len(hi.Members))
	h.UnregisterChan <- c
	hi, err = HubInfoByID("www")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(hi.Members))
}
