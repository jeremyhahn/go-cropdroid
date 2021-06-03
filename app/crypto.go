package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Crypto struct {
	key []byte
}

func NewCrypto(key []byte) *Crypto {
	return &Crypto{key: key}
}

func (crypto *Crypto) Encrypt(data []byte, encode bool) ([]byte, error) {
	block, err := aes.NewCipher(crypto.key)
	if err != nil {
		panic(err)
	}
	b := data
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte(""), err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], b)
	if encode {
		ciphertext = []byte(base64.StdEncoding.EncodeToString(ciphertext))
	}
	return ciphertext, nil
}

func (crypto *Crypto) Decrypt(encrypted []byte, decode bool) ([]byte, error) {
	data := encrypted
	block, err := aes.NewCipher(crypto.key)
	if err != nil {
		panic(err)
	}
	if decode {
		d, err := base64.StdEncoding.DecodeString(string(encrypted))
		if err != nil {
			return []byte(""), err
		}
		data = d
	}
	if len(data) < aes.BlockSize {
		return []byte(""), fmt.Errorf("Data less than block size: %d", aes.BlockSize)
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)
	return data, nil
}
