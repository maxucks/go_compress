package compress

import "bytes"

type Compressor interface {
	Compress(values []int) (*bytes.Buffer, error)
	Decompress(buf *bytes.Buffer) ([]int, error)
}

type Encoder interface {
	EncodeInt(buf *bytes.Buffer, num int) error
	DecodeInt(buf *bytes.Buffer) (int, error)
}
