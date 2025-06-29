package compress

import (
	"bytes"
	"errors"
)

type VLQEncoder struct{}

func (s *VLQEncoder) EncodeInt(buf *bytes.Buffer, num int) error {
	for {
		b := byte(num & 0x7F)
		num >>= 7
		if num == 0 {
			buf.WriteByte(b)
			return nil
		}
		buf.WriteByte(0x80 | b)
	}
}

func (s *VLQEncoder) DecodeInt(buf *bytes.Buffer) (int, error) {
	var result int
	var shift uint

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}

		// Take 7 bits from byte
		result |= int(b&0x7F) << shift

		// If MSB is not set, this is the last byte
		if (b & 0x80) == 0 {
			break
		}

		shift += 7
		if shift > 35 {
			return 0, errors.New("VLQ decoding overflow")
		}
	}

	return result, nil
}
