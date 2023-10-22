package room

import (
	"encoding/json"
	"errors"
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

func testJoinClient(wsURL string, joinURL string) (*websocket.Conn, error) {
	resp, err := http.Post(joinURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tokenPair := map[string]string{
		"token":   "",
		"refresh": "",
	}
	err = json.Unmarshal(body, &tokenPair)
	if err != nil {
		return nil, err
	}

	if tokenPair["token"] == "" {
		return nil, errors.New("invalid token")
	}
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	c.WriteJSON(map[string]interface{}{
		"action": "register_client",
		"data": map[string]string{
			"token": tokenPair["token"],
		},
	})
	return c, err
}

func TestListHub(t *testing.T) {
	roomName := "adsf"
	router := gin.Default()
	RegisterRouter(router.Group("/api/room"), hub{sigSrv: signaling.New()})
	s := httptest.NewServer(router)
	defer s.Close()
	joinU := url.URL{Scheme: "http", Host: strings.TrimPrefix(s.URL, "http://"), Path: "/api/room/" + roomName + "/"}
	u := url.URL{Scheme: "ws", Host: strings.TrimPrefix(s.URL, "http://"), Path: "/api/room/" + roomName + "/ws/"}
	c1, err := testJoinClient(u.String(), joinU.String())
	assert.Nil(t, err)
	c2, err := testJoinClient(u.String(), joinU.String())
	assert.Nil(t, err)
	c1.ReadJSON(nil) // TODO verify
	c2.ReadJSON(nil) // TODO verify

	resp, err := http.Get(s.URL + "/api/room/" + roomName + "/")
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	info := signaling.RoomInfo{}
	assert.Nil(t, json.Unmarshal(body, &info))
	assert.Equal(t, roomName, info.ID)
	assert.Equal(t, 2, len(info.Members))
}
