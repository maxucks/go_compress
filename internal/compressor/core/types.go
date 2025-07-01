package core

import "bytes"

type Compressor interface {
	Compress(values []byte) (*bytes.Buffer, error)
	Decompress(buf *bytes.Buffer) ([]byte, error)
}

type Encoder interface {
	EncodeInt(buf *bytes.Buffer, num int) error
	DecodeInt(buf *bytes.Buffer) (int, error)
}
