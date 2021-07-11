package signaling

import (
	"encoding/json"
	"log"
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
	Clients        map[string]*Client
	RegisterChan   chan *Client // must be unbuffered chan to make sure send await register
	UnregisterChan chan *Client // must be unbuffered chan

	simpleChan     chan *simpleData
	simpleJSONChan chan *simpleJSONData
}

func CreateHub() *Hub {
	h := &Hub{
		Clients:        map[string]*Client{},
		RegisterChan:   make(chan *Client),
		UnregisterChan: make(chan *Client),

		simpleChan:     make(chan *simpleData, 8192),
		simpleJSONChan: make(chan *simpleJSONData, 8192),
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
		case dat := <-h.simpleChan:
			if _, ok := h.Clients[dat.toID]; ok {
				h.Clients[dat.toID].send <- dat.data
			}
		case dat := <-h.simpleJSONChan:
			h.Clients[dat.toID].sendJSON <- dat.data
		}
	}
}
