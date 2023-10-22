package signaling

import "errors"

// RoomList list all rooms
func (s *server) RoomList() []string {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()
	roomsID := []string{}
	for c := range s.rooms {
		roomsID = append(roomsID, c)
	}
	return roomsID
}

// RoomInfoByID returns RoomInfo by roomID
func (s *server) RoomInfoByID(roomID string) (*RoomInfo, error) {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()
	r, ok := s.rooms[roomID]
	if !ok {
		return nil, errors.New("room not found")
	}
	datChan := make(chan *RoomInfo, 1)
	r.RequestInfoChan <- &datChan
	roomDat := <-datChan
	if roomDat == nil {
		return nil, errors.New("room has been closed")
	}
	return roomDat, nil
}
