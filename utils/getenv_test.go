package utils_test

import (
	"os"
	"testing"

	"github.com/ray1422/YAM-api/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) { // for coverage LOL
	s := "ASDF"
	s2 := "qwerty"
	os.Setenv("ASDF", s)
	assert.Equal(t, s, utils.GetEnv("ASDF", "QWERTY"))
	assert.Equal(t, s2, utils.GetEnv("QWERTY123456", s2))
}
