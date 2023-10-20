package signaling

import "errors"

// HubList list all hubs
func HubList() []string {
	globalHubsLock.RLock()
	hubsID := []string{}
	for c := range hubs {
		hubsID = append(hubsID, c)
	}
	globalHubsLock.RUnlock()
	return hubsID
}

// HubInfoByID returns HubInfo
func HubInfoByID(hubID string) (*HubInfo, error) {
	globalHubsLock.RLock()
	defer globalHubsLock.RUnlock()
	h, ok := hubs[hubID]
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
