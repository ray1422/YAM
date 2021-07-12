package signaling

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

type FORWARD_DATA_TYPE int32

const (
	OFFER     FORWARD_DATA_TYPE = iota
	ANSWER    FORWARD_DATA_TYPE = iota
	CANDIDATE FORWARD_DATA_TYPE = iota
)

type Client struct {
	hub      *Hub
	id       string
	conn     *websocket.Conn
	send     chan []byte
	sendJSON chan interface{}
}

func (h *Hub) NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		hub:      h,
		id:       xid.New().String(),
		conn:     conn,
		send:     make(chan []byte, 256),
		sendJSON: make(chan interface{}),
	}

	if conn != nil {
		go client.ReadLoop()
		go client.WriteLoop()
	} else {
		log.Println("conn is nil, it should only happen in units test.")
	}

	return client
}
func (c *Client) close() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.send)
	close(c.sendJSON)
}

func (c *Client) ReadLoop() {
	defer func() {
		c.hub.UnregisterChan <- c
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		action, data, err := parseAction(msg)
		if err != nil {
			continue
		}
		switch action {
		case "register_client":
			c.registerClient(data)

		case "provide_offer":
			if dat, err := c.provideData(data, OFFER); err == nil {
				c.hub.simpleChan <- dat
			}

		case "provide_answer":
			if dat, err := c.provideData(data, ANSWER); err == nil {
				c.hub.simpleChan <- dat
			}
		case "provide_candidate":
			if dat, err := c.provideData(data, CANDIDATE); err == nil {
				c.hub.simpleChan <- dat
			}
		}
	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.hub.UnregisterChan <- c
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			writer.Write(msg)
			err = writer.Close()
			if err != nil {
				return
			}
		case msg, ok := <-c.sendJSON:
			if !ok {
				// chan is closed then close the conn
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteJSON(msg)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type ForwardData struct {
	RemoteID string `json:"remote_id"`
	Data     string `json:"data"`
}
type RegisterData struct {
	Token string `json:"token"`
}

func (c *Client) registerClient(data json.RawMessage) error {
	obj := RegisterData{}
	err := json.Unmarshal(data, &obj)
	token := obj.Token
	_ = token
	// TODO verify token
	if err != nil {
		return err
	}
	log.Println(c.hub.RegisterChan)
	c.hub.RegisterChan <- c // then hub will send list_client action to the client
	return nil
}

func (c *Client) provideData(rawData json.RawMessage, provideType FORWARD_DATA_TYPE) (*simpleData, error) {
	dat := ForwardData{Data: ""}
	err := json.Unmarshal(rawData, &dat)
	if err != nil || dat.Data == "" {
		return nil, errors.New("bad request from client " + c.id)
	}
	toID := dat.RemoteID
	forwardDat := dat
	forwardDat.RemoteID = c.id // keeps the data but change the ID
	actionName := ""
	switch provideType {
	case OFFER:
		actionName = "forward_offer"
	case ANSWER:
		actionName = "forward_answer"
	case CANDIDATE:
		actionName = "forward_candidate"
	default:
		return nil, errors.New("unknown action from client " + c.id)
	}
	var dataStr json.RawMessage

	dataStr, err = json.Marshal(&forwardDat)
	if err != nil {
		log.Println(err)
		return nil, errors.New("something went wrong")
	}
	actionObj := ActionWrapper{Action: actionName, Data: dataStr}
	actionStr, err := json.Marshal(actionObj)
	if err != nil {
		log.Println(err)
		return nil, errors.New("something went wrong")
	}
	return &simpleData{data: actionStr, toID: toID}, nil
}

type ActionWrapper struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// return: action, data, err
func parseAction(inp []byte) (string, json.RawMessage, error) {
	obj := ActionWrapper{Action: ""}
	err := json.Unmarshal(inp, &obj)
	if err != nil {
		return "", nil, err
	} else if obj.Action == "" {
		return "", nil, errors.New("bad request")
	}
	return obj.Action, obj.Data, nil
}
