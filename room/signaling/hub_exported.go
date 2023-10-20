package signaling

import "errors"

// RoomList list all hubs
func (s *server) RoomList() []string {
	s.hubLock.RLock()
	defer s.hubLock.RUnlock()
	hubsID := []string{}
	for c := range s.hubs {
		hubsID = append(hubsID, c)
	}
	return hubsID
}

// RoomInfoByID returns HubInfo
func (s *server) RoomInfoByID(hubID string) (*HubInfo, error) {
	s.hubLock.RLock()
	defer s.hubLock.RUnlock()
	h, ok := s.hubs[hubID]
	if !ok {
		return nil, errors.New("hub not found")
	}
	datChan := make(chan *HubInfo, 1)
	h.RequestInfoChan <- &datChan
	hubDat := <-datChan
	if hubDat == nil {
		return nil, errors.New("hub has been closed")
	}
	return hubDat, nil
}
