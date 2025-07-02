package encoder

import (
	"bytes"
	"fmt"
)

type CompactVLQ struct{}

func (e *CompactVLQ) EncodeInt(buf *bytes.Buffer, num int) error {
	if num < 0 {
		return fmt.Errorf("CompactVLQ: negative values not supported")
	}

	n := num
	for {
		b := byte(n % 128)
		n /= 128
		if n != 0 {
			b |= 0x80
		}
		if err := buf.WriteByte(b); err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}

	return nil
}

func (e *CompactVLQ) DecodeInt(buf *bytes.Buffer) (int, error) {
	var result int
	var shift uint

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("CompactVLQ: failed to read byte: %w", err)
		}
		result += int(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift > 63 {
			return 0, fmt.Errorf("CompactVLQ: value too large")
		}
	}

	return result, nil
}
