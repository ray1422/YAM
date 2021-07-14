package signaling

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	Rooms []Hub
)

type simpleData struct {
	fromID string
	toID   string
	data   []byte
}
type simpleJSONData struct {
	fromID string
	toID   string
	data   interface{}
}
type Hub struct {
	ID             string
	Clients        map[string]*Client
	RegisterChan   chan *Client // must be unbuffered chan to make sure send await register
	UnregisterChan chan *Client // must be unbuffered chan

	simpleChan     chan *simpleData
	simpleJSONChan chan *simpleJSONData

	cleanTimer time.Timer
}

var (
	cleanerTimeout = 120 * time.Second
)

func CreateHub(roomID string) *Hub {
	h := &Hub{
		ID:             roomID,
		Clients:        map[string]*Client{},
		RegisterChan:   make(chan *Client),
		UnregisterChan: make(chan *Client),
		simpleChan:     make(chan *simpleData, 8192),
		simpleJSONChan: make(chan *simpleJSONData, 8192),
		cleanTimer:     *time.NewTimer(cleanerTimeout),
	}
	go h.HubLoop()
	return h
}

// loop for hub. should be create in goroutine
func (h *Hub) HubLoop() {
	defer func() {
		log.Println("hub closed")
	}()
	for {
		select {
		case client := <-h.RegisterChan:
			if !h.cleanTimer.Stop() { // hub has been closed
				fmt.Println("hub has been closed")
				select {
				case <-h.cleanTimer.C:
				default:
				}

				client.hubClosed <- true
				client.close()
				return
			}
			h.cleanTimer.Reset(cleanerTimeout)
			clientsID := []string{}
			for i := range h.Clients {
				clientsID = append(clientsID, i)
			}
			clientsIDBytes, err := json.Marshal(map[string]interface{}{"clients": clientsID, "self_client_id": client.id})
			if err != nil {
				log.Println("something went wrong", err)
				continue
			}
			action := ActionWrapper{Action: "list_client", Data: json.RawMessage(clientsIDBytes)}
			clientListBytes, err := json.Marshal(action)
			if err != nil {
				log.Println("something went wrong")
				continue
			}
			client.send <- clientListBytes
			h.Clients[client.id] = client
		case client := <-h.UnregisterChan:
			if client == nil || h.Clients[client.id] == nil {
				continue
			}
			client.close()
			delete(h.Clients, client.id)
			h.cleanTimer.Reset(5 * time.Second)
		case dat := <-h.simpleChan:
			if _, ok := h.Clients[dat.toID]; ok {
				h.Clients[dat.toID].send <- dat.data
			}
		case dat := <-h.simpleJSONChan:
			h.Clients[dat.toID].sendJSON <- dat.data

		case <-h.cleanTimer.C:
			for _, c := range h.Clients {
				c.hubClosed <- true
			}
			globalHubsLock.Lock()
			delete(hubs, h.ID)
			globalHubsLock.Unlock()
			return
		}

	}
}
