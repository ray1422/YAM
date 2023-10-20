package room

import (
	"github.com/ray1422/YAM-api/room/signaling"
)

func hubList() []string {
	return signaling.HubList()
}

func hubInfo(hubID string) (*signaling.HubInfo, error) {
	return signaling.HubInfoByID(hubID)
}
