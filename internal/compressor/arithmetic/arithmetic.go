package arithmetic

import (
	"bytes"
	"io"
	"math/big"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
	"github.com/maxucks/go_compress.git/internal/compressor/encoder"
)

var (
	MIN_RANGE = float64(0.0)
	MAX_RANGE = float64(1.0)
)

// Arithmetic Coding Compressor
// Unfortunatelly this compressor should be improved
// as I didn't handle how to make it compress data with good results
type ArithmeticCompressor struct {
	cfg     *config
	encoder core.Encoder
}

func NewCompressor(options ...ArithmeticCompressorOption) *ArithmeticCompressor {
	cfg := defaultConfig()
	cfg.apply(options)

	return &ArithmeticCompressor{
		encoder: &encoder.VLQEncoder{},
		cfg:     cfg,
	}
}

func (c *ArithmeticCompressor) Compress(data []byte) (*bytes.Buffer, error) {
	dataLeft, chunksCount := len(data), 0
	var buf bytes.Buffer

	for dataLeft > 0 {
		chunkSize := c.cfg.chunkSize
		if chunkSize > dataLeft {
			chunkSize = dataLeft
		}

		start := chunksCount * c.cfg.chunkSize
		chunk := append([]byte{}, data[start:start+chunkSize]...)

		compressed, err := c.compressChunk(chunk)
		if err != nil {
			return nil, err
		}
		chunkBytes := compressed.Bytes()

		if err := c.encoder.EncodeInt(&buf, len(chunkBytes)); err != nil {
			return nil, err
		}
		buf.Write(chunkBytes)

		dataLeft -= chunkSize
		chunksCount++
	}

	return &buf, nil
}

func (c *ArithmeticCompressor) compressChunk(data []byte) (*bytes.Buffer, error) {
	frmap := c.computeFrequencyMap(data)
	pbmap := c.frequencyToProbabilityMap(frmap)
	value := c.computeValue(data, pbmap)

	return c.encode(value, len(data), frmap)
}

func (c *ArithmeticCompressor) Decompress(buf *bytes.Buffer) ([]byte, error) {
	out := make([]byte, 0)

	for {
		chunkSize, err := c.encoder.DecodeInt(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		chunk := make([]byte, chunkSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return nil, err
		}

		chunkBytes, err := c.decompressChunk(bytes.NewBuffer(chunk[:n]))
		if err != nil {
			return nil, err
		}

		out = append(out, chunkBytes...)
	}

	return out, nil
}

func (c *ArithmeticCompressor) decompressChunk(buf *bytes.Buffer) ([]byte, error) {
	value, numsCount, frmap, err := c.decode(buf)
	if err != nil {
		return nil, err
	}

	data := make([]byte, 0, numsCount)
	pbmap := c.frequencyToProbabilityMap(frmap)

	low, high := c.float(MIN_RANGE), c.float(MAX_RANGE)

	for i := 0; i < numsCount; i++ {
		r := c.emptyFloat().Sub(high, low)

		for sym, pb := range pbmap {
			l := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.low))
			h := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.high))

			if value.Cmp(l) >= 0 && value.Cmp(h) < 0 {
				data = append(data, byte(sym))
				low, high = l, h
				break
			}
		}
	}

	return data, nil
}

// Computes the frequencies of the unique numbers of the source array
func (c *ArithmeticCompressor) computeFrequencyMap(data []byte) frequencyMap {
	fr := make(frequencyMap)
	for _, sym := range data {
		fr[sym]++
	}
	return fr
}

func (c *ArithmeticCompressor) frequencyToProbabilityMap(frmap frequencyMap) ProbabilityMap {
	keys := make([]int, 0, len(frmap))
	total := float64(0)

	for k, p := range frmap {
		total += float64(p)
		keys = append(keys, int(k))
	}

	// Order matters. If iteration order during computing probability map differs at encode
	// and decode stages it leads to completely different ranges breaking decoding
	sort.Ints(keys)

	pbmap := make(ProbabilityMap, len(frmap))
	r := c.float(MIN_RANGE)

	for i, sym := range keys {
		b := byte(sym)
		fr := float64(frmap[b])
		pb := c.emptyFloat().Quo(c.float(fr), c.float(total))

		low, high := c.emptyFloat().Copy(r), c.emptyFloat().Add(r, pb)
		if i == len(keys)-1 {
			high = c.float(MAX_RANGE)
		}

		pbmap[b] = probability{low, high}
		r.Add(r, pb)
	}

	return pbmap
}

func (c *ArithmeticCompressor) computeValue(data []byte, pbmap ProbabilityMap) *big.Float {
	low, high := c.float(MIN_RANGE), c.float(MAX_RANGE)

	for _, sym := range data {
		pb := pbmap[byte(sym)]

		r := c.emptyFloat().Sub(high, low)
		l := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.low))
		h := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.high))

		low, high = l, h
	}

	// Pick the middle of the range as any in [low; high] is allowed
	value := c.emptyFloat().Add(low, high)
	value.Quo(value, c.float(2))

	return value
}

func (c *ArithmeticCompressor) encode(value *big.Float, numsCount int, frmap frequencyMap) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := c.encodeMeta(&buf, numsCount, frmap); err != nil {
		return nil, err
	}
	c.encodeValue(&buf, value)
	return &buf, nil
}

func (c *ArithmeticCompressor) decode(buf *bytes.Buffer) (*big.Float, int, frequencyMap, error) {
	numsCount, frmap, err := c.decodeMeta(buf)
	if err != nil {
		return nil, 0, nil, err
	}

	value, err := c.decodeValue(buf)
	if err != nil {
		return nil, 0, nil, err
	}

	return value, numsCount, frmap, nil
}

func (c *ArithmeticCompressor) encodeValue(buf *bytes.Buffer, value *big.Float) error {
	bytes, err := value.GobEncode()
	if err != nil {
		return err
	}
	c.encoder.EncodeInt(buf, len(bytes))
	_, err = buf.Write(bytes)
	return err
}

func (c *ArithmeticCompressor) decodeValue(buf *bytes.Buffer) (*big.Float, error) {
	valueBytesCount, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, valueBytesCount)
	if _, err := buf.Read(bytes); err != nil {
		return nil, err
	}

	value := c.emptyFloat()
	if err := value.GobDecode(bytes); err != nil {
		return nil, err
	}

	return value, nil
}

// Serializes metadata: length of source array and frequency map
func (c *ArithmeticCompressor) encodeMeta(buf *bytes.Buffer, numsCount int, frmap frequencyMap) error {
	toEncode := make([]int, 0, len(frmap)*2+2)
	toEncode = append(toEncode, numsCount, len(frmap))

	for num, fr := range frmap {
		toEncode = append(toEncode, int(num), fr)
	}

	for _, num := range toEncode {
		if err := c.encoder.EncodeInt(buf, num); err != nil {
			return err
		}
	}

	return nil
}

func (c *ArithmeticCompressor) decodeMeta(buf *bytes.Buffer) (int, frequencyMap, error) {
	numsCount, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return 0, nil, err
	}

	frmapLen, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return 0, nil, err
	}

	frmap := make(frequencyMap, frmapLen)

	for range frmapLen {
		sym, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return 0, nil, err
		}
		fr, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return 0, nil, err
		}
		frmap[byte(sym)] = fr
	}

	return numsCount, frmap, nil
}

func (c *ArithmeticCompressor) float(x float64) *big.Float {
	return new(big.Float).SetPrec(c.cfg.precision).SetFloat64(x)
}

func (c *ArithmeticCompressor) emptyFloat() *big.Float {
	return new(big.Float).SetPrec(c.cfg.precision)
}
