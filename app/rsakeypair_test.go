package app

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRsaEncryptionWithoutCompression(t *testing.T) {

	// 190 char max length!
	secret := "GDE5vXzr2l6fF8udB9hmrWQcdvhRTnXkbDTON+FzqSgWubS1YNb1+i0ju8xgqwyCGXBLmh7qsHUIkZX45ogBnc2xii4ru9Ko2ujo0GR8aTVvnccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	kp, err := CreateRsaKeyPair(NewUnitTestContext().Logger, "../keys", rsa.PSSSaltLengthAuto)
	assert.Nil(t, err)

	encrypted, err := kp.Encrypt(secret, false)
	assert.Nil(t, err)

	//println(encrypted)

	decrypted, err := kp.Decrypt(encrypted, false)
	assert.Nil(t, err)
	assert.Equal(t, string(decrypted), secret)
}

func TestRsaEncryptionWithCompression(t *testing.T) {

	secret := "GDE5vXzr2l6fF8udB9hmrWQcdvhRTnXkbDTON+FzqSgWubS1YNb1+i0ju8xgqwyCGXBLmh7qsHUIkZX45ogBnc2xii4ru9Ko2ujo0GR8aTVvnccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	kp, err := CreateRsaKeyPair(NewUnitTestContext().Logger, "../keys", rsa.PSSSaltLengthAuto)
	assert.Nil(t, err)

	encrypted, err := kp.Encrypt(secret, true)
	assert.Nil(t, err)

	//println(encrypted)

	decrypted, err := kp.Decrypt(encrypted, true)
	assert.Nil(t, err)
	assert.Equal(t, decrypted, secret)
}
