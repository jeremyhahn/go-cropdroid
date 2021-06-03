package test

import (
	"testing"

	"github.com/lni/goutils/stringutil"
	"github.com/stretchr/testify/assert"
)

func TestIpAddressWithPort(t *testing.T) {
	assert.True(t, stringutil.IsValidAddress("192.168.1.1:8000"))
}

func TestIpAddressWithoutPort(t *testing.T) {
	assert.False(t, stringutil.IsValidAddress("192.168.1.1"))
}

func TestLocalDnsWithPort(t *testing.T) {
	assert.True(t, stringutil.IsValidAddress("cropdroid-0.cropdroid:8000"))
}

func TestLocalDnsWithoutPort(t *testing.T) {
	assert.False(t, stringutil.IsValidAddress("cropdroid-0.cropdroid"))
}
