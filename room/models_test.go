package room

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/ray1422/YAM-api/room/signaling"
	"github.com/stretchr/testify/assert"
)

func TestListRoom(t *testing.T) {
	// create a hub
	signaling.GlobalHubsLock.Lock()
	signaling.Hubs["www"] = signaling.CreateHub("www")
	signaling.Hubs["asdf"] = signaling.CreateHub("asdf")
	signaling.GlobalHubsLock.Unlock()
	ss := hubList()
	assert.ElementsMatch(t, []string{"www", "asdf"}, ss)
}

func TestRoomInfo(t *testing.T) {
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
	signaling.GlobalHubsLock.Lock()
	signaling.Hubs["www"] = signaling.CreateHub("www")
	signaling.Hubs["asdf"] = signaling.CreateHub("asdf")
	signaling.GlobalHubsLock.Unlock()
	h := signaling.Hubs["www"]
	c := h.NewClient(ws)
	h.RegisterChan <- c
	hi, err := hubInfo("www")
	assert.Nil(t, err)
	assert.Equal(t, hi.ID, "www")
	assert.Equal(t, 1, len(hi.Members))
	h.UnregisterChan <- c
	hi, err = hubInfo("www")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(hi.Members))

}

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}
