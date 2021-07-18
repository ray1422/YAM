package room

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ray1422/YAM-api/room/signaling"
	"github.com/stretchr/testify/assert"
)

func testJoinClient(url string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	c.WriteJSON(map[string]interface{}{
		"action": "register_client",
		"data": map[string]string{
			"token": "asdf",
		},
	})
	return c, err
}

func TestListHub(t *testing.T) {
	roomName := "adsf"
	router := gin.Default()
	RegisterRouter(router.Group("/api/room"))
	s := httptest.NewServer(router)
	defer s.Close()
	u := url.URL{Scheme: "ws", Host: strings.TrimPrefix(s.URL, "http://"), Path: "/api/room/" + roomName + "/ws/"}
	c1, _ := testJoinClient(u.String())
	c2, _ := testJoinClient(u.String())
	c1.ReadJSON(nil) // TODO verify
	c2.ReadJSON(nil) // TODO verify

	resp, err := http.Get(s.URL + "/api/room/" + roomName + "/")
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	info := signaling.HubInfo{}
	assert.Nil(t, json.Unmarshal(body, &info))
	assert.Equal(t, roomName, info.ID)
	assert.Equal(t, 2, len(info.Members))
}
