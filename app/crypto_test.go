package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {

	crypto := NewCrypto([]byte("thisis32bitlongpassphraseimusing"))

	secret := "!secret!"

	ciphertext, err := crypto.Encrypt([]byte(secret), false)
	assert.Nil(t, err)

	decrypted, err := crypto.Decrypt(ciphertext, false)
	assert.Nil(t, err)

	assert.Equal(t, secret, string(decrypted))
}

func TestCryptoEncoding(t *testing.T) {

	crypto := NewCrypto([]byte("thisis32bitlongpassphraseimusing"))

	secret := "!secret!"

	ciphertext, err := crypto.Encrypt([]byte(secret), true)
	assert.Nil(t, err)

	println(ciphertext)

	decrypted, err := crypto.Decrypt(ciphertext, true)
	assert.Nil(t, err)

	assert.Equal(t, secret, string(decrypted))
}
