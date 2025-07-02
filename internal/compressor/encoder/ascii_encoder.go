package encoder

import (
	"bytes"
	"errors"

	"github.com/maxucks/go_compress.git/internal/compressor/utils"
)

const ASCII_MAX_VAL = 127

// This encoder is my own packing bits implementation
// Packs any number (up to 16256) representing as either 1 or 2 bytes of ASCII code into buffer
// It's pretty restricted and I used it for testing
type ASCII struct{}

func (s *ASCII) EncodeInt(buf *bytes.Buffer, num int) error {
	chunksCount := num / ASCII_MAX_VAL
	if chunksCount > ASCII_MAX_VAL {
		return errors.New("too big number")
	}

	if chunksCount > 0 {
		b0 := 1<<7 | byte(chunksCount)
		buf.WriteByte(b0)
	}

	b1 := byte(num - ASCII_MAX_VAL*chunksCount)
	buf.WriteByte(b1)

	return nil
}

func (s *ASCII) DecodeInt(buf *bytes.Buffer) (int, error) {
	var num, chunksCount int

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}

		isLastChunk := !utils.IsNthBitSet(b, 7)

		if isLastChunk {
			num = int(b)
			break
		} else {
			chunksCount = int(b & ASCII_MAX_VAL)
		}
	}

	num += ASCII_MAX_VAL * chunksCount
	return num, nil
}
