package signaling

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ray1422/YAM-api/utils/jwt"
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
	token := jwt.New(48 * time.Hour)
	token.Payload["room_id"] = "www"
	assert.Nil(t, c1.registerClient([]byte(`{"token": "`+token.TokenString()+`"}`)))
	assert.Nil(t, c2.registerClient([]byte(`{"token": "`+token.TokenString()+`"}`)))

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
	os.Setenv("DEBUG", "true")
	roomName := "neo"
	done := make(chan bool)
	router := gin.Default()
	var c1ID string
	RoomWS(router.Group("/api/room"), "/")
	s := httptest.NewServer(router)
	defer s.Close()

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
	// defer c1.Close() // c1 will be closed in the following section to test the leave signal
	c2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	msg := "yet_another_data+w"
	defer c2.Close()

	// readloop for c1
	go func(t *testing.T, msg string) {
		token := jwt.New(48 * time.Hour)
		token.Payload["room_id"] = roomName
		c1.WriteMessage(websocket.TextMessage, []byte(`{"action": "register_client", "data": {"token": "`+token.TokenString()+`"}}`))
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
					c2ID := <-c2IDChan
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
		token := jwt.New(48 * time.Hour)
		token.Payload["room_id"] = roomName
		c2.WriteMessage(websocket.TextMessage, []byte(`{"action": "register_client", "data": {"token": "`+token.TokenString()+`"}}`))
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
					c2IDChan <- obj.SelfClientID
				} else {
					assert.Equal(t, 1, len(obj.Clients))
					c1ID = <-c1IDChan
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
	c1.Close()
	leaveSig := struct {
		Action string `json:"action"`
		Data   struct {
			RemoteID string `json:"remote_id"`
			Event    string `json:"event"`
		}
	}{}
	assert.Nil(t, c2.ReadJSON(&leaveSig))
	assert.Equal(t, "client_event", leaveSig.Action)
	assert.NotEmpty(t, leaveSig.Data.RemoteID)
	assert.Equal(t, "leave", leaveSig.Data.Event)
	ch := make(chan *HubInfo, 1)

	Hubs[roomName].RequestInfoChan <- &ch

	// time.Sleep(1 * time.Second)
	// GlobalHubsLock.RLock()
	// h, ok := Hubs[roomName]
	info := <-ch
	assert.Equal(t, 1, len(info.Members))
	// GlobalHubsLock.Unlock()

}

func newWS() *websocket.Conn {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()
	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, _ := websocket.DefaultDialer.Dial(u, nil)

	return ws
}

func echo(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
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
