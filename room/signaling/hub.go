package signaling

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
	for {
		select {
		case client := <-h.RegisterChan:
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
