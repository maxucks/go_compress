package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compressor/encoder"
	"github.com/maxucks/go_compress.git/internal/compressor/mixed"
	"github.com/maxucks/go_compress.git/internal/compressor/utils"
)

// func ComputeOptimalChunkSize(inputSize uint, avgProbability float64, maxPrecisionBits uint) (uint, error) {
// 	if avgProbability <= 0 || avgProbability >= 1 {
// 		return 0, errors.New("avgProbability must be in (0, 1)")
// 	}

// 	bitsPerSymbol := math.Log2(1.0 / avgProbability)

// 	maxChunkSize := uint(float64(maxPrecisionBits) / bitsPerSymbol)
// 	if maxChunkSize > inputSize {
// 		maxChunkSize = inputSize
// 	}

// 	return maxChunkSize, nil
// }

func SlicesEqual[T comparable](a, b []T) bool {
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

// TODO: move to default compressor
func serializeInts(ints []int) ([]byte, error) {
	encoder := encoder.VLQEncoder{}
	var buf bytes.Buffer
	for _, num := range ints {
		err := encoder.EncodeInt(&buf, num)
		// err := binary.Write(&buf, binary.BigEndian, uint16(num))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// TODO: move to default compressor
func deserializeInts(data []byte) ([]int, error) {
	encoder := encoder.VLQEncoder{}

	ints := make([]int, 0)
	buf := bytes.NewBuffer(data)

	for {
		num, err := encoder.DecodeInt(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ints = append(ints, int(num))
	}

	return ints, nil
}

func PrintTable(inputSize int, uniqueSymbolsCount uint) {
	bitsPerSymbol := math.Log2(float64(uniqueSymbolsCount))

	for chunkSize := 10; chunkSize < inputSize; chunkSize += 10 {
		precision := math.Ceil(float64(chunkSize) * bitsPerSymbol)
		fmt.Printf("chunkSize = %v, precision = %v+\n", chunkSize, precision)
	}
}

func PickOptimalPrecisionAndChunkSize(inputSize int, uniqueSymbolsCount uint, maxPrecision int) (uint, uint, error) {
	bitsPerSymbol := math.Log2(float64(uniqueSymbolsCount))

	// Try max possible chunk size under precision cap
	chunkSize := int(float64(maxPrecision) / bitsPerSymbol)
	if chunkSize > inputSize {
		chunkSize = inputSize
	}

	precision := math.Ceil(float64(chunkSize) * bitsPerSymbol)

	return uint(chunkSize), uint(precision), nil
}

const maxValue = 300
const numsCount = 1000

func readLogfile() []byte {
	data, err := os.ReadFile("logfile.txt")
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func saveLogfile(data []byte) {
	file, err := os.OpenFile("logfile.compressed", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// chunkSize, precision := 500, uint(3492) // 26.899999999999995%
	// chunkSize, precision := 250, uint(1465) // 34.99999999999999%
	// chunkSize, precision := 200, uint(1100) // 35.25%
	// chunkSize, precision := 125, uint(598) // 37.199999999999996%
	// chunkSize, precision := 100, uint(460) // 40.5%
	// chunkSize, precision := 100, uint(442) // 40.5%
	// chunkSize, precision := uint(50), uint(166) // 37%
	// chunkSize, precision := 25, uint(57) // 30%
	// chunkSize, precision := 20, uint(35) // 22.5%
	// chunkSize, precision := 10, uint(35) // 22.5%

	chunkSize, precision := 100, uint(500)

	fmt.Printf("Run with chunkSize = %v, precision = %v\n", chunkSize, precision)

	nums := utils.SeedByRatio(numsCount, maxValue, 0.95)
	// nums := utils.SeedRandom(numsCount, maxValue)
	sort.Ints(nums)

	// for i := 0; i < len(nums)-50; i += 50 {
	// 	fmt.Printf("%v = %v\n", i, nums[i:i+50])
	// }

	data, _ := serializeInts(nums)

	// data := readLogfile()

	compressor := mixed.NewCompressor()

	fmt.Printf("-- COMPRESSION --\n")
	compressed, err := compressor.Compress(data)
	if err != nil {
		log.Fatalf("failed to compress: %s", err)
	}

	compressedBytes := compressed.Bytes()
	saveLogfile(compressedBytes)

	for i := 0; i < len(compressedBytes)-50; i += 50 {
		fmt.Printf("%v = %v\n", i, compressedBytes[i:i+50])
	}

	fmt.Printf("-- DECOMPRESSION --\n")
	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		log.Fatalf("failed to decompress: %s", err)
	}

	fmt.Println("DONE")
	// fmt.Println(decompressed)

	fmt.Println("input = ", data)
	fmt.Println("decompressed = ", decompressed)

	if !SlicesEqual(data, decompressed) {
		log.Fatal("slices are not equal")
	}

	originalSize := len(data)
	compressedSize := len(compressedBytes)
	compression := (1 - float64(compressedSize)/float64(originalSize)) * 100

	fmt.Printf("compression: %v%%\n", compression)
}
