package signaling

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type simpleData struct {
	toID string
	data []byte
}

type Addr struct {
	Network string
	Address string
}
type Member struct {
	Addr Addr   `json:"addr"`
	ID   string `json:"id"`
}

type HubInfo struct {
	ID      string
	Members []Member
}

// Hub hub
type Hub struct {
	ID             string
	Clients        map[string]*Client
	RegisterChan   chan *Client // must be unbuffered chan to make sure send await register
	UnregisterChan chan *Client // must be unbuffered chan

	simpleChan chan *simpleData

	RequestInfoChan chan *chan *HubInfo

	cleanTimer time.Timer
}

var (
	cleanerTimeout = 120 * time.Second
)

// CreateHub CreateHub
func CreateHub(roomID string) *Hub {
	h := &Hub{
		ID:              roomID,
		Clients:         map[string]*Client{},
		RegisterChan:    make(chan *Client),
		UnregisterChan:  make(chan *Client),
		simpleChan:      make(chan *simpleData, 8192),
		RequestInfoChan: make(chan *chan *HubInfo, 512),
		cleanTimer:      *time.NewTimer(cleanerTimeout),
	}
	go h.HubLoop()
	return h
}

// HubLoop loop for hub. should be create in goroutine
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
			action := actionWrapper{Action: "list_client", Data: json.RawMessage(clientsIDBytes)}
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

		case ch := <-h.RequestInfoChan:
			ms := []Member{}
			for idx, c := range h.Clients {
				ms = append(ms, Member{
					ID: idx,
					Addr: Addr{
						Network: c.conn.RemoteAddr().Network(),
						Address: c.conn.RemoteAddr().String(),
					},
				})
			}
			*ch <- &HubInfo{
				ID:      h.ID,
				Members: ms,
			}
		case <-h.cleanTimer.C:
			for _, c := range h.Clients {
				c.hubClosed <- true
			}
			select {
			case ch := <-h.RequestInfoChan:
				*ch <- nil
			default:
			}
			GlobalHubsLock.Lock()
			delete(Hubs, h.ID)
			GlobalHubsLock.Unlock()
			return
		}

	}
}
