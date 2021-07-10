package signaling

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiClients(t *testing.T) {
	hub := CreateHub()
	c1 := hub.NewClient(nil)
	c2 := hub.NewClient(nil)
	assert.NotNil(t, c1, c2)
	dat, err := c1.provideData([]byte(`{"remote_id":"`+c2.id+`", "data": "yet_another_data"}`), OFFER)
	assert.Equal(t, c2.id, dat.toID)
	assert.Nil(t, err)
	c1.hub.simpleChan <- dat
	receiveBytes := <-c2.send
	aw := ActionWrapper{}
	err = json.Unmarshal(receiveBytes, &aw)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "forward_offer", aw.Action)
	fw := ForwardData{}
	err = json.Unmarshal(aw.Data, &fw)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "yet_another_data", fw.Data)
}
