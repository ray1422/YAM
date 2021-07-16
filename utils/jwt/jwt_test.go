package jwt

import (
	"log"
	"testing"
)

func TestJWT(t *testing.T) {
	jwt := JWT{
		header: map[string]string{
			"alg": "HS256",
			"typ": "jwt",
		},
		payload: map[string]string{
			"asdf": "adsf_fff",
		},
	}
	log.Println(jwt.genUnencryptedPartStr())
}
