package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHasher(t *testing.T) {

	password := "$ecret!"
	passwordHasher := NewPasswordHasher()
	hash, err := passwordHasher.Encrypt(password)
	assert.Nil(t, err)
	assert.NotNil(t, hash)

	match, err := passwordHasher.Compare(password, hash)
	assert.Nil(t, err)
	assert.True(t, match)

	fmt.Println(hash)
}
