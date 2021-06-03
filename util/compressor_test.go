package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressor(t *testing.T) {
	compressor := NewCompressor()
	data := "this is a really long string that takes up a lot of space"

	compressed, err := compressor.Zip([]byte(data))
	assert.Nil(t, err)

	decompressed, err := compressor.Unzip(compressed)
	assert.Nil(t, err)

	assert.Equal(t, data, string(decompressed))
}
