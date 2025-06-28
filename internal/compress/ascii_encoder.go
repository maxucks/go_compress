package compress

import (
	"bytes"
	"errors"

	"github.com/maxucks/go_compress.git/internal/utils"
)

const ASCII_MAX_VAL = 127

type ASCIIEncoder struct{}

// Packs any number (up to 16256) representing as either 1 or 2 bytes of ASCII code into buffer
func (s *ASCIIEncoder) EncodeInt(buf *bytes.Buffer, num int) error {
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

// Unpacks first n bytes representing an integer
func (s *ASCIIEncoder) DecodeInt(buf *bytes.Buffer) (int, error) {
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
