package signaling

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseAction(t *testing.T) {
	_, _, err := parseAction([]byte("{}"))
	assert.NotNil(t, err)
	_, _, err = parseAction([]byte("{"))
	assert.NotNil(t, err)
	action, _, err := parseAction([]byte(`{"action":     "provide_answer"}`))
	if assert.Nil(t, err) {
		assert.Equal(t, "provide_answer", action)
	}

}
func TestClientProvideDat(t *testing.T) {
	client := CreateHub("www").NewClient(nil)
	fmt.Println("client ID:", client.id)
	{
		_, err := client.provideData([]byte(`{"remote_id":"www"}`), Offer)
		assert.NotNil(t, err, "")
	}
	{
		_, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), 1000)
		assert.NotNil(t, err, "")
	}
	// test for offer
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), Offer)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := actionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_offer", string(actObj.Action))
		}
		obj := forwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	// test for answer
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), Answer)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := actionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_answer", string(actObj.Action))
		}
		obj := forwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	// test for candidate
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), Candidate)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := actionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_candidate", string(actObj.Action))
		}
		obj := forwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	client.hub.UnregisterChan <- client
}

func TestClientRegisterTimeout(t *testing.T) {
	h := CreateHub("asdf")
	c := h.NewClient(newWS())
	time.Sleep(5500 * time.Millisecond)
	assert.NotNil(t, c.conn.WriteJSON(&map[string]string{}))
}
func TestClientRegisterHubClosed(t *testing.T) {
	h := CreateHub("asdf")
	c := h.NewClient(newWS())
	h.cleanTicker.Reset(1)
	time.Sleep(5500 * time.Millisecond)
	assert.NotNil(t, c.conn.WriteJSON(&map[string]string{}))
}
