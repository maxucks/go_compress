package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compress"
	"github.com/maxucks/go_compress.git/internal/utils"
)

func serializeInts(ints []int) ([]byte, error) {
	var buf bytes.Buffer
	for _, num := range ints {
		err := binary.Write(&buf, binary.BigEndian, uint16(num))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TODO: with small numsCount everything works fine, but with large it glitches
func main() {
	const maxValue = 300
	const numsCount = 4

	nums := utils.SeedNumbers(numsCount, maxValue)
	sort.Ints(nums)

	options := []compress.ArithmeticCompressorOption{
		// compress.WithMetaCompression,
	}

	var compressor compress.Compressor = compress.NewArithmeticCompressor(options...)
	buf, err := compressor.Compress(nums)
	if err != nil {
		log.Fatalf("failed to compress: %s", err)
	}

	compressed := buf.Bytes()

	decompressedNums, err := compressor.Decompress(buf)
	if err != nil {
		log.Fatalf("failed to decompress: %s", err)
	}

	sort.Ints(decompressedNums)

	fmt.Printf("nums = %v\n", nums)
	fmt.Printf("deco = %v\n", decompressedNums)

	rawbytes, _ := serializeInts(nums)
	fmt.Printf("size without compression = %vb\n", len(rawbytes))
	fmt.Printf("size after compression = %vb\n", len(compressed))
	fmt.Printf("compression = %v%%\n", 100-100/float64(len(rawbytes))*float64(len(compressed)))

	if !slicesEqual(nums, decompressedNums) {
		log.Fatal("slices are not equal")
	}
}
