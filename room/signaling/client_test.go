package signaling

import (
	"encoding/json"
	"fmt"
	"testing"

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
	client := CreateHub().NewClient(nil)
	fmt.Println("client ID:", client.id)
	{
		_, err := client.provideData([]byte(`{"remote_id":"www"}`), OFFER)
		assert.NotNil(t, err, "")
	}
	{
		_, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), 1000)
		assert.NotNil(t, err, "")
	}
	// test for offer
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), OFFER)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := ActionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_offer", string(actObj.Action))
		}
		obj := ForwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	// test for answer
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), ANSWER)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := ActionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_answer", string(actObj.Action))
		}
		obj := ForwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	// test for candidate
	{
		dat, err := client.provideData([]byte(`{"remote_id":"www", "data": "yet_another_data"}`), CANDIDATE)
		if assert.Nil(t, err) {
			assert.Equal(t, "www", dat.toID)
		}
		actObj := ActionWrapper{}
		err = json.Unmarshal(dat.data, &actObj)
		if assert.Nil(t, err) {
			assert.Equal(t, "forward_candidate", string(actObj.Action))
		}
		obj := ForwardData{}
		err = json.Unmarshal(actObj.Data, &obj)
		if assert.Nil(t, err) {
			assert.Equal(t, "yet_another_data", string(obj.Data))
			assert.Equal(t, client.id, obj.RemoteID)
		}
	}
	client.hub.UnregisterChan <- client
}
