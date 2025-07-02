package mixed

import (
	"bytes"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
	"github.com/maxucks/go_compress.git/internal/compressor/haffman"
	"github.com/maxucks/go_compress.git/internal/compressor/lz77"
)

type MixedCompressor struct {
	main        core.Compressor
	postprocess core.Compressor
}

func NewCompressor() *MixedCompressor {
	return &MixedCompressor{
		main: lz77.NewCompressor(
			lz77.SetBuffer(64),
		),
		postprocess: haffman.NewCompressor(),
	}
}

func (c *MixedCompressor) Compress(data []byte) (*bytes.Buffer, error) {
	compressed, err := c.main.Compress(data)
	if err != nil {
		return nil, err
	}

	return c.postprocess.Compress(compressed.Bytes())
}

func (c *MixedCompressor) Decompress(buf *bytes.Buffer) ([]byte, error) {
	decompressed, err := c.postprocess.Decompress(buf)
	if err != nil {
		return nil, err
	}

	return c.main.Decompress(bytes.NewBuffer(decompressed))
}
