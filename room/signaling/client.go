package signaling

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ray1422/YAM-api/utils/jwt"
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
	maxMessageSize = 1048576
)

// ForwardDataType int32
type ForwardDataType int32

const (
	// Offer Offer
	Offer ForwardDataType = iota
	// Answer Answer
	Answer ForwardDataType = iota
	// Candidate Candidate
	Candidate ForwardDataType = iota
)

// Client client for each ws connection, handling read and write loop
type Client struct {
	hub       *Hub
	id        string
	conn      *websocket.Conn
	send      chan []byte
	hubClosed chan bool
	register  chan bool
}

// NewClient new client
func (h *Hub) NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		hub:       h,
		id:        xid.New().String(),
		conn:      conn,
		send:      make(chan []byte, 256),
		hubClosed: make(chan bool, 1),
		register:  make(chan bool, 1),
	}

	if conn != nil {
		go client.readLoop()
		go client.writeLoop()
		t := time.NewTimer(5 * time.Second)
		go func() {
			select {
			case <-t.C:
				select {
				case client.hub.UnregisterChan <- client:
				case <-client.hubClosed:
				}
				log.Println("waiting for registering timeout")
			case <-client.register:
				close(client.register)
				client.register = nil
				t.Stop()
			}
		}()

	} else {
		log.Println("conn is nil, it should only happen in units test.")
	}

	return client
}
func (c *Client) close() {
	if c.conn != nil {
		log.Println("close client")
		c.conn.Close()
	}
	// close(c.send)
	// close(c.sendJSON)
}

func (c *Client) readLoop() {
	defer func() {
		select {
		case c.hub.UnregisterChan <- c:
		case <-c.hubClosed:
		}
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
			if dat, err := c.provideData(data, Offer); err == nil {
				c.hub.simpleChan <- dat
			}

		case "provide_answer":
			if dat, err := c.provideData(data, Answer); err == nil {
				c.hub.simpleChan <- dat
			}
		case "provide_candidate":
			if dat, err := c.provideData(data, Candidate); err == nil {
				c.hub.simpleChan <- dat
			}
		}
	}
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		select {
		case c.hub.UnregisterChan <- c:
		case <-c.hubClosed:
		}

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

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type forwardData struct {
	RemoteID string `json:"remote_id"`
	Data     string `json:"data"`
}
type registerData struct {
	Token string `json:"token"`
}

func (c *Client) registerClient(data json.RawMessage) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = errors.New("client has been registered")
		}
	}()
	if c.register == nil {
		return errors.New("client has been registered")
	}

	obj := registerData{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	token := obj.Token
	jwt, err := jwt.FromString(token)
	if err != nil {
		return err
	}
	if !jwt.Check() {
		log.Println("invalid token:", token)
		return errors.New("invalid token")
	}
	if v, ok := jwt.Payload["room_id"]; ok && v == c.hub.ID {

	} else {
		return errors.New("invalid token")
	}
	timer := time.NewTimer(cleanerTimeout)

	select {
	case c.hub.RegisterChan <- c: // then hub will send list_client action to the client
	case <-timer.C:
		return errors.New("register timeout")
	}
	c.register <- true
	return nil
}

func (c *Client) provideData(rawData json.RawMessage, provideType ForwardDataType) (*simpleData, error) {
	dat := forwardData{Data: ""}
	err := json.Unmarshal(rawData, &dat)
	if err != nil || dat.Data == "" {
		return nil, errors.New("bad request from client " + c.id)
	}
	toID := dat.RemoteID
	forwardDat := dat
	forwardDat.RemoteID = c.id // keeps the data but change the ID
	actionName := ""
	switch provideType {
	case Offer:
		actionName = "forward_offer"
	case Answer:
		actionName = "forward_answer"
	case Candidate:
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
	actionObj := actionWrapper{Action: actionName, Data: dataStr}
	actionStr, err := json.Marshal(actionObj)
	if err != nil {
		log.Println(err)
		return nil, errors.New("something went wrong")
	}
	return &simpleData{data: actionStr, toID: toID}, nil
}

type actionWrapper struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// return: action, data, err
func parseAction(inp []byte) (string, json.RawMessage, error) {
	obj := actionWrapper{Action: ""}
	err := json.Unmarshal(inp, &obj)
	if err != nil {
		return "", nil, err
	} else if obj.Action == "" {
		return "", nil, errors.New("bad request")
	}
	return obj.Action, obj.Data, nil
}
