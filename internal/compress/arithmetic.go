package compress

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/maxucks/go_compress.git/internal/utils"
)

var (
	MIN_RANGE = float64(0.0)
	MAX_RANGE = float64(1.0)
)

// Arithmetic Coding Compressor
type ArithmeticCompressor struct {
	cfg     *config
	encoder Encoder
}

func NewArithmeticCompressor(options ...ArithmeticCompressorOption) *ArithmeticCompressor {
	cfg := defaultConfig()
	cfg.apply(options)

	return &ArithmeticCompressor{
		encoder: &VLQEncoder{},
		cfg:     cfg,
	}
}

func (c *ArithmeticCompressor) Compress(data []int) (*bytes.Buffer, error) {
	return c.compress(data, c.cfg.compressMeta)
}

func (c *ArithmeticCompressor) Decompress(buf *bytes.Buffer) ([]int, error) {
	return c.decompress(buf, c.cfg.compressMeta)
}

func (c *ArithmeticCompressor) compress(data []int, compressMeta bool) (*bytes.Buffer, error) {
	frmap := c.computeFrequencyMap(data)
	pbmap := c.frequencyToProbabilityMap(frmap)
	value := c.computeValue(data, pbmap)

	return c.encode(value, len(data), frmap, compressMeta)
}

func (c *ArithmeticCompressor) decompress(buf *bytes.Buffer, decompressMeta bool) ([]int, error) {
	value, numsCount, frmap, err := c.decode(buf, decompressMeta)
	if err != nil {
		return nil, err
	}

	data := make([]int, 0, numsCount)
	pbmap := c.frequencyToProbabilityMap(frmap)

	low, high := c.float(MIN_RANGE), c.float(MAX_RANGE)

	for i := 0; i < numsCount; i++ {
		r := c.emptyFloat().Sub(high, low)

		for sym, pb := range pbmap {
			l := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.low))
			h := c.emptyFloat().Add(low, c.emptyFloat().Mul(r, pb.high))

			if value.Cmp(l) >= 0 && value.Cmp(h) < 0 {
				data = append(data, sym)
				low, high = l, h
				break
			}
		}
	}

	return data, nil
}

// Computes the frequencies of the unique numbers of the source array
func (c *ArithmeticCompressor) computeFrequencyMap(data []int) frequencyMap {
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
		fr := float64(frmap[sym])
		pb := c.emptyFloat().Quo(c.float(fr), c.float(total))

		low, high := c.emptyFloat().Copy(r), c.emptyFloat().Add(r, pb)
		if i == len(keys)-1 {
			high = c.float(MAX_RANGE)
		}

		pbmap[sym] = probability{low, high}
		r.Add(r, pb)
	}

	return pbmap
}

func (c *ArithmeticCompressor) computeValue(data []int, pbmap ProbabilityMap) *big.Float {
	low, high := c.float(MIN_RANGE), c.float(MAX_RANGE)

	for _, sym := range data {
		pb := pbmap[sym]

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

func (c *ArithmeticCompressor) encode(value *big.Float, numsCount int, frmap frequencyMap, compressMeta bool) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := c.encodeMeta(&buf, numsCount, frmap); err != nil {
		return nil, err
	}

	if compressMeta {
		encodedMetaBytes := buf.Bytes()
		buf.Reset()

		if err := c.compressMeta(&buf, encodedMetaBytes); err != nil {
			return nil, err
		}
	}

	c.encodeValue(&buf, value)
	return &buf, nil
}

func (c *ArithmeticCompressor) decode(buf *bytes.Buffer, decompressMeta bool) (*big.Float, int, frequencyMap, error) {
	if decompressMeta {
		if err := c.decompressMeta(buf); err != nil {
			return nil, 0, nil, err
		}
	}

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
		frmap[sym] = fr
	}

	return numsCount, frmap, nil
}

// TODO: remove
func (c *ArithmeticCompressor) compressMeta(buf *bytes.Buffer, bytes []byte) error {
	metadata := utils.BytesToInts(bytes)

	// meta should be compressed only once
	compressedMetaBuf, err := c.compress(metadata, false)
	if err != nil {
		return err
	}
	compressedMetaBytes := compressedMetaBuf.Bytes()

	c.encoder.EncodeInt(buf, len(compressedMetaBytes))
	buf.Write(compressedMetaBytes)

	return nil
}

// TODO: remove
func (c *ArithmeticCompressor) decompressMeta(buf *bytes.Buffer) error {
	metaBytesCount, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return err
	}

	metaBytes := make([]byte, metaBytesCount)

	n, err := buf.Read(metaBytes)
	if err != nil {
		return err
	}
	if n != metaBytesCount {
		return fmt.Errorf("failed to decompress meta: read %d meta bytes, but expected %d", n, metaBytesCount)
	}

	var metaBuffer bytes.Buffer
	metaBuffer.Write(metaBytes)

	// meta should be decompressed only once
	encodedMeta, err := c.decompress(&metaBuffer, false)
	if err != nil {
		return err
	}

	encodedMetaBytes := utils.IntsToBytes(encodedMeta)

	// Puts meta bytes to the biginning of the buffer
	restBytes := buf.Bytes()
	buf.Reset()

	buf.Write(encodedMetaBytes)
	buf.Write(restBytes)

	return nil
}

func (c *ArithmeticCompressor) float(x float64) *big.Float {
	return new(big.Float).SetPrec(c.cfg.precision).SetFloat64(x)
}

func (c *ArithmeticCompressor) emptyFloat() *big.Float {
	return new(big.Float).SetPrec(c.cfg.precision)
}
