package lz77

import (
	"bytes"
	"io"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
	"github.com/maxucks/go_compress.git/internal/compressor/encoder"
	"github.com/maxucks/go_compress.git/internal/compressor/utils"
)

type LZ77Compressor struct {
	cfg     *config
	encoder core.Encoder
}

func NewCompressor(options ...LZ77CompressorOption) *LZ77Compressor {
	cfg := defaultConfig()
	cfg.apply(options)

	return &LZ77Compressor{
		cfg:     cfg,
		encoder: &encoder.VLQEncoder{},
	}
}

func (c *LZ77Compressor) Compress(data []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	for wend := 0; wend < len(data); {
		wstart := utils.Max(0, wend-c.cfg.windowSize)
		pos := c.findMaxSublistPosition(data, wstart, wend)

		var next byte

		if pos == nil || pos.count == 0 {
			pos = &position{0, 0}
			next = data[wend]
			wend++
		} else {
			idx := wend + pos.count
			if idx < len(data) {
				next = data[idx]
				wend += pos.count + 1
			} else {
				next = 0
				wend += pos.count
			}
		}

		if err := c.encodeLink(&buf, pos.offset, pos.count, next); err != nil {
			return nil, err
		}
	}

	return &buf, nil
}

func (c *LZ77Compressor) Decompress(buf *bytes.Buffer) ([]byte, error) {
	bytes := make([]byte, 0)

	for {
		offset, count, next, err := c.decodeLink(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if offset == 0 && count == 0 {
			bytes = append(bytes, next)
		} else {
			start := len(bytes) - offset
			for i := 0; i < count; i++ {
				bytes = append(bytes, bytes[start+i])
			}
			// Treat next == 0 as nil
			if next != 0 {
				bytes = append(bytes, next)
			}
		}
	}

	return bytes, nil
}

// Treat next = 0 as nil
func (c *LZ77Compressor) encodeLink(buf *bytes.Buffer, offset, count int, next byte) error {
	if err := c.encoder.EncodeInt(buf, offset); err != nil {
		return err
	}
	if err := c.encoder.EncodeInt(buf, count); err != nil {
		return err
	}

	return c.encoder.EncodeInt(buf, int(next))
}

// Treat next = 0 as nil
func (c *LZ77Compressor) decodeLink(buf *bytes.Buffer) (int, int, byte, error) {
	offset, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return 0, 0, 0, err
	}
	count, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return 0, 0, 0, err
	}
	next, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return 0, 0, 0, err
	}

	return offset, count, byte(next), nil
}

func (c *LZ77Compressor) findMaxSublistPosition(data []byte, wstart, wend int) *position {
	offset, length := 0, 0
	var pos *position

	window := data[wstart:wend]

	for size := 1; size <= len(data)-wend; size++ {
		target := data[wend : wend+size]
		idx := utils.FindLastIndex(window, target)
		if idx == -1 {
			break
		}
		offset, length = len(window)-idx, len(target)
		pos = &position{offset, length}
	}

	return pos
}
