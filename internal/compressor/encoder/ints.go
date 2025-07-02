package encoder

import (
	"bytes"
	"io"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
)

// Non-standart encoder which aim is to convert input numbers into bytes
type Ints struct {
	encoder core.Encoder
}

func NewIntsEncoder() *Ints {
	return &Ints{
		encoder: &CompactVLQ{},
	}
}

func (s *Ints) Encode(ints []int) ([]byte, error) {
	var buf bytes.Buffer
	for _, num := range ints {
		err := s.encoder.EncodeInt(&buf, num)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (s *Ints) Decode(data []byte) ([]int, error) {
	ints := make([]int, 0)
	buf := bytes.NewBuffer(data)

	for {
		num, err := s.encoder.DecodeInt(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ints = append(ints, int(num))
	}

	return ints, nil
}
