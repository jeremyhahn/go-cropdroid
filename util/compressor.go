package util

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
)

type GZipCompressor struct {
}

func NewCompressor() *GZipCompressor {
	return &GZipCompressor{}
}

func (compressor *GZipCompressor) Zip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	gz.Flush()
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (compressor *GZipCompressor) Unzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return []byte(""), err
	}
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	io.Copy(os.Stdout, r)
	return d, nil
}
