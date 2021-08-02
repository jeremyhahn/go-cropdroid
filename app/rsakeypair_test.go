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

func TestUserDefinedRsaKeyPair(t *testing.T) {

	// 190 char max length!
	secret := "GDE5vXzr2l6fF8udB9hmrWQcdvhRTnXkbDTON+FzqSgWubS1YNb1+i0ju8xgqwyCGXBLmh7qsHUIkZX45ogBnc2xii4ru9Ko2ujo0GR8aTVvnccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	kp, err := CreateUserDefinedRsaKeyPair(NewUnitTestContext().Logger,
		"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEArwt2ym+MAFvAb6cAXl7uAZEKJ5m6StSspbyhnhUTqyoCa7BK\nYe4lDwW356epkZ5Hyl6lXCqT8XxGaU+pDBXUB1l0i7QSOxVv/SfUYanEKmNISW51\nuu+0n+4J5IG3uyRX3XiRItsOsbjL894jJyEchNXYdWGHu9EnK2Ytc6G9LJO1AiC5\nIPzWusBMH7ijth/z30ZnL7W4nbJqNK50bNr/HA2nQsl07n8/guOvLZbV+v0TI1wd\nwD6c/Tarj89UCYLrK5YwWfcifLIqtoL9XTXwjUmLQk/xYIqcEEmEsFrgZBmF6ht1\nTwEVEa0tX8GiMLZPNqImPMgaTm6+QLY6wGDr1wIDAQABAoIBACUdgSqbTEwnKD6E\nYoegCUc7wbNz3RRQ0+qwfHRQc8MvPSQoVR0+qYzt4Xi1DDdcIEzAlL9eJ9BkUWmz\nAl0Vo8eLKDMXE2aDvSModtfeb0Gtm342dbAVc28VwfM5rgN4SUkkb2G8oPj9/gDP\ncRSy6KEh1qvM6kLgrjV9jNWfzcTNktnyGmld/a/POMFySk+tp7D3LFNpQt0qCNsq\nV8Vg8qBAxI4O/kNbQ0nIv9DQlDrUseYME6XBOmigTxFogMucZ1/udX4+4dDIE7en\nDBiHPIk1YbaZ/WpK2urBRgg1RTDexBegcPvMMkd7YwSfxVEpM2yLqYPnwpYkQxd2\nOV7aAUECgYEA5a88q9Nq2XQSFgtk/CkZKpCWCGYfpV4X1KrVXHd8WEu2IUSHm8PI\n0Gkwd8bxQ4YHF6MIg37DSgIoh/SGrf/tDfPdOFvl/yM5WX5w61F8/m1yFVhIuIkV\nC1B8aeVABdqaNaMAcLMNBIwBUMtqq4Mys2+Xo2OFwI8Re6Pzb6e3VQsCgYEAwxmd\nq9IO1o43jzt5UI9inbVJZrLJZTiWMv4ShJVDrk7hFphaFiOcOafZEFD5fG8yLIdK\nbwofzD2fBREyNMX9RTA+/C4IS0u7Dj1TSoAXeIb5AnEbokmG3ECI2lxcfGUrNxr0\nilpGCyYQeftOrKzzrj3Yu3NlTruNucZMSLfrK+UCgYB8FfYDJk7dd/WlbzZ5fIKa\nGk7T7sg+AN2DCWAHeo307cJRqsJQhq9g2NNUgmgpgKkoPe4FjGBZBV18RcDVFCSv\nmwXywsM42YDMNqEuoHGUyvANVArFl1mFKVBtrWqvPvB89bjxKepogHLdgWf5jQHB\nKxKTNNs3spNRZrvHoKZNDQKBgClCrZm83u55PT7JcSqcaFq6ED/r57PEd99o5Dmt\n55ZhkDDbH5I3Db8TxFAzD9BFI/NO2WsKVRc4oPzNWjTW+m07etaSVaa26WRli8vh\nsxUGVnsxuIplymOiMk8b4WNdcfpBdR4dYVrSPgHOKCFUomRjKAbcrLwt5hc33MI0\nQ0QRAoGBAN58/DcjwefnDdRzzhmUlSfiAMMDkkNLV54nok9awbmWDIovtMHGb7gu\nrd4O7gDcF7NJr2Oe4wWoLoL0t5yWGxZ7WEI9mupbMuYbaz8u0bpTKm7yc5yNwUOQ\nvdhrfydS4kCEQQ7XYzJwcjys8h7WQjGfTK8tcQ9R/oEsGA9b2UdY\n-----END RSA PRIVATE KEY-----\n",
		"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArwt2ym+MAFvAb6cAXl7u\nAZEKJ5m6StSspbyhnhUTqyoCa7BKYe4lDwW356epkZ5Hyl6lXCqT8XxGaU+pDBXU\nB1l0i7QSOxVv/SfUYanEKmNISW51uu+0n+4J5IG3uyRX3XiRItsOsbjL894jJyEc\nhNXYdWGHu9EnK2Ytc6G9LJO1AiC5IPzWusBMH7ijth/z30ZnL7W4nbJqNK50bNr/\nHA2nQsl07n8/guOvLZbV+v0TI1wdwD6c/Tarj89UCYLrK5YwWfcifLIqtoL9XTXw\njUmLQk/xYIqcEEmEsFrgZBmF6ht1TwEVEa0tX8GiMLZPNqImPMgaTm6+QLY6wGDr\n1wIDAQAB\n-----END PUBLIC KEY-----\n", rsa.PSSSaltLengthAuto)
	assert.Nil(t, err)

	encrypted, err := kp.Encrypt(secret, false)
	assert.Nil(t, err)

	decrypted, err := kp.Decrypt(encrypted, false)
	assert.Nil(t, err)
	assert.Equal(t, string(decrypted), secret)
}
