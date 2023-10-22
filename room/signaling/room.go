package signaling

import (
	"encoding/json"
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

type RoomInfo struct {
	ID      string
	Members []Member
}

// Room room
type Room struct {
	ID             string
	clients        map[string]*Client
	RegisterChan   chan *Client // must be unbuffered chan to make sure send await register
	UnregisterChan chan *Client // must be unbuffered chan

	simpleChan chan *simpleData

	RequestInfoChan chan *chan *RoomInfo

	cleanTicker time.Ticker
}

var (
	cleanerTimeout = 30 * time.Second
)

// RoomCreate RoomCreate
func (s *server) RoomCreate(roomID string) *Room {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()
	if s.rooms[roomID] != nil {
		return s.rooms[roomID]
	}
	h := &Room{
		ID:              roomID,
		clients:         map[string]*Client{},
		RegisterChan:    make(chan *Client),
		UnregisterChan:  make(chan *Client),
		simpleChan:      make(chan *simpleData, 8192),
		RequestInfoChan: make(chan *chan *RoomInfo, 512),
		cleanTicker:     *time.NewTicker(cleanerTimeout),
	}
	s.rooms[roomID] = h
	go func() {
		h.Loop()
		s.roomsLock.Lock()
		delete(s.rooms, h.ID)
		s.roomsLock.Unlock()
	}()
	return h
}

// Loop loop for room. should be create in goroutine
func (r *Room) Loop() {
	defer func() {
		log.Println("room closed")
	}()
	for {
		select {
		case client := <-r.RegisterChan:
			r.cleanTicker.Reset(cleanerTimeout)
			clientsID := []string{}
			for i := range r.clients {
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
			r.clients[client.id] = client
		case client := <-r.UnregisterChan:
			client.close()
			delete(r.clients, client.id)
			r.cleanTicker.Reset(5 * time.Second)
			for _, c := range r.clients {
				b, err := json.Marshal(map[string]interface{}{
					"action": "client_event",
					"data": map[string]string{
						"remote_id": client.id,
						"event":     "leave",
					},
				})
				if err == nil {
					c.send <- b
				} else {
					log.Println(err)
				}
			}
		case dat := <-r.simpleChan:
			if _, ok := r.clients[dat.toID]; ok {
				r.clients[dat.toID].send <- dat.data
			}

		case ch := <-r.RequestInfoChan:
			ms := []Member{}
			for idx, c := range r.clients {
				ms = append(ms, Member{
					ID: idx,
					Addr: Addr{
						Network: c.conn.RemoteAddr().Network(),
						Address: c.conn.RemoteAddr().String(),
					},
				})
			}
			*ch <- &RoomInfo{
				ID:      r.ID,
				Members: ms,
			}
		case <-r.cleanTicker.C:
			if len(r.clients) > 0 {
				r.cleanTicker.Reset(cleanerTimeout)
				continue
			}
			log.Println("room " + r.ID + " close due to no clients in room")
			for _, c := range r.clients {
				c.roomClosed <- true
			}
			select {
			case ch := <-r.RequestInfoChan:
				*ch <- nil
			default:
			}
			return
		}

	}
}
