package gzip

import (
	"bytes"
	"compress/gzip"
)

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	wr, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return data, err
	}

	_, err = wr.Write(data)
	if err != nil {
		return data, err
	}

	err = wr.Close()
	if err != nil {
		return data, err
	}

	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
