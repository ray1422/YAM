package room

import (
	"errors"

	"github.com/ray1422/YAM-api/room/signaling"
)

func hubList() []string {
	signaling.GlobalHubsLock.RLock()
	hubsID := []string{}
	for c := range signaling.Hubs {
		hubsID = append(hubsID, c)
	}
	signaling.GlobalHubsLock.RUnlock()
	return hubsID
}

func hubInfo(hubID string) (*signaling.HubInfo, error) {
	signaling.GlobalHubsLock.RLock()
	defer signaling.GlobalHubsLock.RUnlock()
	h, ok := signaling.Hubs[hubID]
	if !ok {
		return nil, errors.New("hub not found")
	}
	datChan := make(chan *signaling.HubInfo, 1)
	h.RequestInfoChan <- &datChan
	hubDat := <-datChan
	if hubDat == nil {
		return nil, errors.New("hub has been closed")
	}
	return hubDat, nil
}
