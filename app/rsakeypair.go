package app

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io/ioutil"

	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"

	jwt "github.com/dgrijalva/jwt-go"
)

type KeyPair interface {
	GetDirectory() string
	GetPrivateKey() *rsa.PrivateKey
	GetPrivateBytes() []byte
	GetPublicKey() *rsa.PublicKey
	GetPublicBytes() []byte
	Encrypt(message string, compress bool) (string, error)
	Decrypt(base64CipherText string, decompress bool) (string, error)
	Sign(message []byte, saltLen int) (crypto.Hash, []byte, []byte, error)
	Verify(messageSHA crypto.Hash, hashed []byte, signature []byte) (bool, error)
}

type RsaKeyPair struct {
	Directory    string
	PrivateKey   *rsa.PrivateKey
	PrivateBytes []byte
	PublicKey    *rsa.PublicKey
	PublicBytes  []byte
	PSSOptions   *rsa.PSSOptions
	SHA256       hash.Hash
	label        []byte
	KeyPair
}

func NewRsaKeyPair(logger *logging.Logger, keydir string) (KeyPair, error) {
	return CreateRsaKeyPair(logger, keydir, rsa.PSSSaltLengthAuto)
}

func CreateRsaKeyPair(logger *logging.Logger, directory string, saltLen int) (KeyPair, error) {
	logger.Debugf("[RsaKeyPair] Loading key files from %s", directory)
	privateKeyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "rsa.key"))
	if err != nil {
		logger.Errorf("[RsaKeyPair] %s", err.Error())
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		logger.Errorf("[RsaKeyPair] %s", err.Error())
		return nil, err
	}
	publicKeyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "rsa.pub"))
	if err != nil {
		logger.Errorf("[RsaKeyPair] %s", err.Error())
		return nil, err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		logger.Errorf("[RsaKeyPair] %s", err.Error())
		return nil, err
	}
	return &RsaKeyPair{
		Directory:    directory,
		PrivateKey:   privateKey,
		PrivateBytes: privateKeyBytes,
		PublicKey:    publicKey,
		PublicBytes:  publicKeyBytes,
		PSSOptions: &rsa.PSSOptions{
			SaltLength: saltLen},
		SHA256: sha256.New()}, nil
}

func (keypair *RsaKeyPair) GetDirectory() string {
	return keypair.Directory
}

func (keypair *RsaKeyPair) GetPrivateKey() *rsa.PrivateKey {
	return keypair.PrivateKey
}

func (keypair *RsaKeyPair) GetPrivateBytes() []byte {
	return keypair.PrivateBytes
}

func (keypair *RsaKeyPair) GetPublicKey() *rsa.PublicKey {
	return keypair.PublicKey
}

func (keypair *RsaKeyPair) GetPublicBytes() []byte {
	return keypair.PublicBytes
}

func (keypair *RsaKeyPair) Encrypt(message string, compress bool) (string, error) {
	compressor := util.NewCompressor()
	ciphertext, err := rsa.EncryptOAEP(keypair.SHA256, rand.Reader, keypair.PublicKey, []byte(message), keypair.label)
	if err != nil {
		return "", err
	}
	if compress {
		bytes, err := compressor.Zip(ciphertext)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(bytes), nil
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (keypair *RsaKeyPair) Decrypt(base64CipherText string, decompress bool) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(base64CipherText)
	if err != nil {
		return "", err
	}
	if decompress {
		compressor := util.NewCompressor()
		d, err := compressor.Unzip(ciphertext)
		if err != nil {
			return "", nil
		}
		ciphertext = d
	}
	plaintext, err := rsa.DecryptOAEP(keypair.SHA256, rand.Reader, keypair.PrivateKey, ciphertext, keypair.label)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (keypair *RsaKeyPair) Sign(message []byte, saltLen int) (crypto.Hash, []byte, []byte, error) {
	messageSHA := crypto.SHA256
	pssh := messageSHA.New()
	pssh.Write(message)
	sum := pssh.Sum(nil)
	signature, err := rsa.SignPSS(rand.Reader, keypair.PrivateKey, messageSHA, sum, keypair.PSSOptions)
	if err != nil {
		return messageSHA, []byte(""), []byte(""), err
	}
	return messageSHA, signature, sum, nil
}

func (keypair *RsaKeyPair) Verify(messageSHA crypto.Hash, hashed []byte, signature []byte) (bool, error) {
	if err := rsa.VerifyPSS(keypair.PublicKey, messageSHA, hashed, signature, keypair.PSSOptions); err != nil {
		return false, err
	}
	return true, nil
}
