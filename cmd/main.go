package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compress"
	"github.com/maxucks/go_compress.git/internal/utils"
)

func ComputeOptimalChunkSize(inputSize uint, avgProbability float64, maxPrecisionBits uint) (uint, error) {
	if avgProbability <= 0 || avgProbability >= 1 {
		return 0, errors.New("avgProbability must be in (0, 1)")
	}

	bitsPerSymbol := math.Log2(1.0 / avgProbability)

	maxChunkSize := uint(float64(maxPrecisionBits) / bitsPerSymbol)
	if maxChunkSize > inputSize {
		maxChunkSize = inputSize
	}

	return maxChunkSize, nil
}

// TODO: move to default compressor
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

// TODO: move to default compressor
func deserializeInts(data []byte) ([]int, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("invalid data length: must be a multiple of 2")
	}

	ints := make([]int, 0, len(data)/2)
	buf := bytes.NewReader(data)

	for buf.Len() > 0 {
		var num uint16
		err := binary.Read(buf, binary.BigEndian, &num)
		if err != nil {
			return nil, err
		}
		ints = append(ints, int(num))
	}

	return ints, nil
}

func PickOptimalPrecisionAndChunkSize(inputSize int, uniqueSymbolsCount uint, maxPrecision int) (uint, uint, error) {
	bitsPerSymbol := math.Log2(float64(uniqueSymbolsCount))

	// Try max possible chunk size under precision cap
	maxChunkSize := uint(float64(maxPrecision) / bitsPerSymbol)
	if maxChunkSize > uint(inputSize) {
		maxChunkSize = uint(inputSize)
	}

	// Recompute actual precision needed for the selected chunk size
	requiredBits := uint(math.Ceil(float64(maxChunkSize) * bitsPerSymbol))

	return maxChunkSize, requiredBits, nil
}

// TODO: with small numsCount everything works fine, but with large it glitches
func main() {
	const maxValue = 300
	const numsCount = 1000
	// const precision = uint(512)

	// avgProbability := 1 / float64(maxValue)
	// optimalChunkSize, _ := ComputeOptimalChunkSize(numsCount, avgProbability, precision)

	chunkSize, precision, err := PickOptimalPrecisionAndChunkSize(numsCount, numsCount, 256)
	if err != nil {
		log.Fatalf("failed to PickOptimalPrecisionAndChunkSize: %s", err)
	}
	fmt.Printf("chunkSize = %v, precision = %v\n", chunkSize, precision)

	nums := utils.SeedNumbers(numsCount, maxValue)
	sort.Ints(nums)

	options := []compress.ArithmeticCompressorOption{
		compress.WithPrecision(precision),
		// compress.WithMetaCompression,
	}

	chunk := make([]int, 0, chunkSize)
	for i := 0; i < int(chunkSize); i++ {
		chunk = append(chunk, nums[i])
	}

	var compressor compress.Compressor = compress.NewArithmeticCompressor(options...)
	buf, err := compressor.Compress(chunk)
	if err != nil {
		log.Fatalf("failed to compress: %s", err)
	}

	compressed := buf.Bytes()

	decompressedChunk, err := compressor.Decompress(buf)
	if err != nil {
		log.Fatalf("failed to decompress: %s", err)
	}
	sort.Ints(decompressedChunk)

	// fmt.Printf("input = %v\n", chunk)
	// fmt.Printf("decompressed = %v\n", decompressedChunk)

	bytes, _ := serializeInts(chunk)
	fmt.Printf("size without compression = %vb\n", len(bytes))
	fmt.Printf("size after compression = %vb\n", len(compressed))
	fmt.Printf("compression = %v%%\n", 100-100/float64(len(bytes))*float64(len(compressed)))

	if !utils.SlicesEqual(chunk, decompressedChunk) {
		log.Fatal("slices are not equal")
	}
}
