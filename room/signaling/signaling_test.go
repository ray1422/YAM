package signaling

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type listClientResponse struct {
	SelfClientID string   `json:"self_client_id"`
	Clients      []string `json:"clients"`
}

func TestMultiClients(t *testing.T) {
	GlobalHubsLock.Lock()
	Hubs["www"] = CreateHub("www")
	hub := Hubs["www"]
	GlobalHubsLock.Unlock()
	c1 := hub.NewClient(nil)
	c2 := hub.NewClient(nil)
	assert.NotNil(t, c1, c2)
	c1.registerClient([]byte(`{"token": "www"}`))
	c2.registerClient([]byte(`{"token": "www" }`))

	// action
	c1ReceiveBytes := <-c1.send
	c2ReceiveBytes := <-c2.send
	aw := actionWrapper{}

	// c1
	assert.Nil(t, json.Unmarshal(c1ReceiveBytes, &aw))
	obj := listClientResponse{}
	json.Unmarshal(aw.Data, &obj)
	assert.Equal(t, 0, len(obj.Clients))
	assert.Equal(t, c1.id, obj.SelfClientID)

	// c2
	assert.Nil(t, json.Unmarshal(c2ReceiveBytes, &aw))
	json.Unmarshal(aw.Data, &obj)
	assert.Equal(t, 1, len(obj.Clients))
	assert.Equal(t, c2.id, obj.SelfClientID)
	assert.Equal(t, c1.id, obj.Clients[0])
	dat, err := c1.provideData([]byte(`{"remote_id":"`+c2.id+`", "data": "yet_another_data"}`), Offer)
	assert.Equal(t, c2.id, dat.toID)
	assert.Nil(t, err)
	c1.hub.simpleChan <- dat
	receiveBytes := <-c2.send
	aw = actionWrapper{}
	err = json.Unmarshal(receiveBytes, &aw)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "forward_offer", aw.Action)
	fw := forwardData{}
	err = json.Unmarshal(aw.Data, &fw)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "yet_another_data", fw.Data)
}

func TestWithRealConn(t *testing.T) {
	roomName := "adsf"
	done := make(chan bool)
	router := gin.Default()
	RoomWS(router.Group("/api/room"), "/")
	s := httptest.NewServer(router)
	defer s.Close()

	// srv := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: router,
	// }

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below

	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: strings.TrimPrefix(s.URL, "http://"), Path: "/api/room/" + roomName + "/ws/"}
	log.Printf("connecting to %s", u.String())
	c1IDChan := make(chan string, 1)
	c2IDChan := make(chan string, 1)
	c1, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c1.Close()
	c2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	msg := "yet_another_data+w"
	defer c2.Close()

	// readloop for c1
	go func(t *testing.T, msg string) {
		c1.WriteMessage(websocket.TextMessage, []byte(`{"action": "register_client", "data": {"token": "w"}}`))
		defer func() {
			fmt.Println("c1 done")
			done <- true
		}()
	READLOOP:
		for {
			_, message, err := c1.ReadMessage()
			log.Printf("c1 recv: %s", message)
			if err != nil {
				log.Println("read:", err)
				return
			}
			action, dat, err := parseAction(message)
			assert.Nil(t, err)
			switch action {
			case "list_client":
				obj := listClientResponse{}
				assert.Nil(t, json.Unmarshal(dat, &obj))
				if len(obj.Clients) == 0 {
					assert.Equal(t, 0, len(obj.Clients))
					assert.NotEqual(t, "", obj.SelfClientID)
					c1IDChan <- obj.SelfClientID
				} else {
					assert.Equal(t, 1, len(obj.Clients))
					c2ID := <-c1IDChan
					assert.Equal(t, c2ID, obj.Clients[0])
					assert.NotEqual(t, "", obj.SelfClientID)
					c1IDChan <- obj.SelfClientID
					c2.NextWriter(websocket.TextMessage)
					err := c1.WriteJSON(&actionWrapper{
						Action: "provide_offer",
						Data:   json.RawMessage(`{"remote_id": "` + c2ID + `", "data": "` + msg + `"}`),
					})
					assert.Nil(t, err)
					break READLOOP
				}

			case "forward_offer":
				c2ID := <-c2IDChan
				obj := forwardData{RemoteID: ""}
				assert.Nil(t, json.Unmarshal(dat, &obj))
				assert.NotEqual(t, "", obj.RemoteID)
				assert.Equal(t, c2ID, obj.RemoteID)
				assert.Equal(t, msg, obj.Data)
				return
			}

		}
	}(t, msg)

	// readloop for c2
	go func(t *testing.T, msg string) {
		c2.WriteMessage(websocket.TextMessage, []byte(`{"action": "register_client", "data": {"token": "w"}}`))
		defer func() {
			fmt.Println("c2 done")
			done <- true
		}()
	READLOOP:
		for {
			_, message, err := c2.ReadMessage()
			log.Printf("c2 recv: %s", message)
			if err != nil {
				log.Println("read:", err)
				return
			}
			action, dat, err := parseAction(message)
			assert.Nil(t, err)
			switch action {
			case "list_client":
				obj := listClientResponse{SelfClientID: ""}
				assert.Nil(t, json.Unmarshal(dat, &obj))
				if len(obj.Clients) == 0 {
					assert.Equal(t, 0, len(obj.Clients))
					assert.NotEqual(t, "", obj.SelfClientID)
					c1IDChan <- obj.SelfClientID
				} else {
					assert.Equal(t, 1, len(obj.Clients))
					c1ID := <-c1IDChan
					assert.Equal(t, c1ID, obj.Clients[0])
					assert.NotEqual(t, "", obj.SelfClientID)
					c2IDChan <- obj.SelfClientID
					c2.NextWriter(websocket.TextMessage)
					err := c2.WriteJSON(&actionWrapper{
						Action: "provide_offer",
						Data:   json.RawMessage(`{"remote_id": "` + c1ID + `", "data": "` + msg + `"}`),
					})
					assert.Nil(t, err)
					break READLOOP
				}
			case "forward_offer":
				c1ID := <-c1IDChan
				obj := forwardData{RemoteID: ""}
				assert.Nil(t, json.Unmarshal(dat, &obj))
				assert.NotEqual(t, "", obj.RemoteID)
				assert.Equal(t, c1ID, obj.RemoteID)
				assert.Equal(t, msg, obj.Data)
				return
			}

		}
	}(t, msg)

	<-done
	<-done
	// TODO get information from hub.
	// c1.Close()
	// time.Sleep(1 * time.Second)
	// h, ok := hubs[roomName]
	// assert.True(t, ok)
	// assert.Equal(t, 1, len(h.Clients))

}
