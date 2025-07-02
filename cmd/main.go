package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/maxucks/go_compress.git/internal/compressor/core"
	"github.com/maxucks/go_compress.git/internal/compressor/encoder"
	"github.com/maxucks/go_compress.git/internal/compressor/haffman"
	"github.com/maxucks/go_compress.git/internal/compressor/lz77"
	"github.com/maxucks/go_compress.git/internal/compressor/mixed"
	"github.com/maxucks/go_compress.git/internal/seed"
)

func Run(testsCount int, seedFn seed.SeedFn, sortInput bool) ([]float64, error) {
	avgCompression := []float64{0, 0, 0}

	compressors := []core.Compressor{
		haffman.NewCompressor(),
		lz77.NewCompressor(),
		mixed.NewCompressor(),
	}

	for range testsCount {
		input := seedFn()
		if sortInput {
			sort.Ints(input)
		}

		for i, cmp := range compressors {
			compression, err := Compress(cmp, input)
			if err != nil {
				return nil, err
			}
			avgCompression[i] += compression
		}
	}

	for i := range len(compressors) {
		avgCompression[i] /= float64(testsCount)
	}

	return avgCompression, nil
}

func Compress(compressor core.Compressor, input []int) (float64, error) {
	encoder := encoder.NewIntsEncoder()
	data, _ := encoder.Encode(input)

	compressed, err := compressor.Compress(data)
	if err != nil {
		return 0, err
	}

	originalSize := len(data)
	compressedSize := len(compressed.Bytes())
	compression := (1 - float64(compressedSize)/float64(originalSize)) * 100

	return compression, nil
}

type seedInfo struct {
	description string
	seedFn      seed.SeedFn
}

func main() {
	testsCount := 250

	for seedType := range 4 {
		for _, size := range []int{50, 100, 500, 1000, 10000} {
			for _, maxValue := range []int{10, 50, 100, 300} {
				if seedType == 0 && maxValue < 50 {
					continue
				}

				var description string
				var seedFn seed.SeedFn

				switch seedType {
				case 0:
					description, seedFn = seed.SeedByRatio(size, maxValue, 0.96)
				case 1:
					description, seedFn = seed.SeedByRatio(size, maxValue, 0.9)
				case 2:
					description, seedFn = seed.SeedByRatio(size, maxValue, 0.7)
				case 3:
					description, seedFn = seed.SeedRandom(size, maxValue)
				}

				compression, err := Run(testsCount, seedFn, true)
				if err != nil {
					log.Fatal(err)
				}

				printRow(compression, description)
			}
		}
	}

	// Edge-case tests

	for _, maxValue := range []int{10, 50, 100, 300} {
		description, seedFn := seed.SeedEqually(3, maxValue)
		compression, err := Run(testsCount, seedFn, false)
		if err != nil {
			log.Fatal(err)
		}
		printRow(compression, description)
	}
}

func printRow(compression []float64, description string) {
	fmt.Printf("| %10.6f%% | %10.6f%% | %10.6f%% | %v |\n", compression[0], compression[1], compression[2], description)
}
