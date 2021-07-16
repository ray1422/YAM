package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	jwt := JWT{
		Header: map[string]string{
			"alg": "HS256",
			"typ": "jwt",
		},
		Payload: map[string]string{
			"expire_time": fmt.Sprint(time.Now().Unix() + 5),
		},
	}
	assert.True(t, jwt.Check())
	delete(jwt.Payload, "expire_time")
	assert.False(t, jwt.Check())
	jwt.Payload["expire_time"] = fmt.Sprint(time.Now().Unix() - 5)
	assert.False(t, jwt.Check())
	s := jwt.TokenString()
	jwt2, err := FromString(s)
	assert.Nil(t, err)
	assert.Equal(t, jwt.GenSignature(), jwt2.GenSignature())

}
