package compress

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/maxucks/go_compress.git/internal/utils"
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
		encoder: &ASCIIEncoder{},
		cfg:     cfg,
	}
}

func (c *ArithmeticCompressor) Compress(values []int) (*bytes.Buffer, error) {
	fmt.Println("-- compression --")
	return c.compress(values, c.cfg.compressMeta)
}

func (c *ArithmeticCompressor) Decompress(buf *bytes.Buffer) ([]int, error) {
	fmt.Println("-- decompression --")
	return c.decompress(buf, c.cfg.compressMeta)
}

func (c *ArithmeticCompressor) compress(values []int, compressMeta bool) (*bytes.Buffer, error) {
	frmap := c.computeFrequencyMap(values)
	pbmap := c.frequencyToProbabilityMap(frmap)
	value := c.computeValue(values, pbmap)

	return c.encode(value, len(values), frmap, compressMeta)
}

func (c *ArithmeticCompressor) decompress(buf *bytes.Buffer, decompressMeta bool) ([]int, error) {
	value, numsCount, frmap, err := c.decode(buf, decompressMeta)
	if err != nil {
		return nil, err
	}

	nums := make([]int, 0, numsCount)
	pbmap := c.frequencyToProbabilityMap(frmap)

	low, high := 0.0, 1.0

	for i := 0; i < numsCount; i++ {
		r := high - low
		for num, pb := range pbmap {
			// Narrowing the range
			l := low + r*pb.low
			h := low + r*pb.high

			if value >= l && value < h {
				// fmt.Printf("search for %v in [%v; %v] => %v\n", value, l, h, num)
				nums = append(nums, num)
				low, high = l, h
				break
			}
		}
	}

	return nums, nil
}

// Computes the frequencies of the unique numbers of the source array
func (c *ArithmeticCompressor) computeFrequencyMap(nums []int) frequencyMap {
	fr := make(frequencyMap)
	for _, num := range nums {
		fr[num]++
	}
	return fr
}

func (c *ArithmeticCompressor) frequencyToProbabilityMap(frmap frequencyMap) ProbabilityMap {
	keys := make([]int, 0, len(frmap))
	total := 0

	for k, p := range frmap {
		total += p
		keys = append(keys, k)
	}

	// Order matters. If iteration order during computing probability map differs at encode
	// and decode stages it leads to completely different ranges breaking decoding
	sort.Ints(keys)

	pbmap := make(ProbabilityMap, len(frmap))
	r := 0.0

	for i, num := range keys {
		frequency := frmap[num]
		pb := float64(frequency) / float64(total)

		low, high := r, r+pb
		if i == len(keys)-1 {
			high = 1.0
		}

		pbmap[num] = probability{pb, low, high}
		r += pb
	}

	fmt.Println(pbmap)

	return pbmap
}

func (c *ArithmeticCompressor) computeValue(values []int, pbmap ProbabilityMap) float64 {
	low, high := 0.0, 1.0

	for _, val := range values {
		pb := pbmap[val]
		r := high - low

		// Narrowing the range
		l := low + r*pb.low
		h := low + r*pb.high
		low, high = l, h
	}

	// As any of the (low; high] is ok then pick one in the middle
	value := (low + high) / 2.0
	// Reducing count of decimal digits to achieve better compression
	value = utils.RoundToPrecision(value, low, high)

	return value
}

func (c *ArithmeticCompressor) encode(value float64, numsCount int, frmap frequencyMap, compressMeta bool) (*bytes.Buffer, error) {
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

func (c *ArithmeticCompressor) decode(buf *bytes.Buffer, decompressMeta bool) (float64, int, frequencyMap, error) {
	if decompressMeta {
		if err := c.decompressMeta(buf); err != nil {
			return 0, 0, nil, err
		}
		fmt.Printf("buffer with decompressed meta = %v\n", buf.Bytes())
	}

	numsCount, frmap, err := c.decodeMeta(buf)
	if err != nil {
		return 0, 0, nil, err
	}

	value, err := c.decodeValue(buf)
	if err != nil {
		return 0, 0, nil, err
	}

	return value, numsCount, frmap, nil
}

// Parses each decimal part's digit of float64 and converts it to the slice of ASCII codes
// For instance 0.01263789455 will be converted to [0 126 37 89 45 5]
func (c *ArithmeticCompressor) encodeValue(buf *bytes.Buffer, value float64) {
	bits := math.Float64bits(value)
	binary.Write(buf, binary.LittleEndian, bits)
}

func (c *ArithmeticCompressor) decodeValue(buf *bytes.Buffer) (float64, error) {
	var bits uint64
	err := binary.Read(buf, binary.LittleEndian, &bits)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// Serializes metadata: length of source array and frequency map
func (c *ArithmeticCompressor) encodeMeta(buf *bytes.Buffer, numsCount int, frmap frequencyMap) error {
	toEncode := make([]int, 0, len(frmap)*2+2)
	toEncode = append(toEncode, numsCount, len(frmap))

	for num, fr := range frmap {
		toEncode = append(toEncode, num, fr)
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
		num, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return 0, nil, err
		}
		fr, err := c.encoder.DecodeInt(buf)
		if err != nil {
			return 0, nil, err
		}
		frmap[num] = fr
	}

	return numsCount, frmap, nil
}

func (c *ArithmeticCompressor) compressMeta(buf *bytes.Buffer, bytes []byte) error {
	fmt.Println("- meta compression -")
	fmt.Printf("meta before compression = %v\n", bytes)
	encodedMeta := utils.BytesToInts(bytes)

	// meta should be compressed only once
	compressedMetaBuf, err := c.compress(encodedMeta, false)
	if err != nil {
		return err
	}
	compressedMetaBytes := compressedMetaBuf.Bytes()

	fmt.Printf("compressedMetaBytes count = %v\n", len(compressedMetaBytes))

	c.encoder.EncodeInt(buf, len(compressedMetaBytes))
	buf.Write(compressedMetaBytes)

	return nil
}

func (c *ArithmeticCompressor) decompressMeta(buf *bytes.Buffer) error {
	fmt.Println("- meta decompression -")
	metaBytesCount, err := c.encoder.DecodeInt(buf)
	if err != nil {
		return err
	}

	fmt.Printf("decoded metaBytesCount count = %v\n", metaBytesCount)

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
