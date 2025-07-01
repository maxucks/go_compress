package utils

import (
	"math/rand"
	"strconv"
	"time"
)

type Frequency struct {
	Num       int
	Frequency int
}

func SeedRandom(count, maxValue int) []int {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	input := make([]int, 0, count)
	for range count {
		input = append(input, rnd.Intn(maxValue-1)+1)
	}

	return input
}

func SeedEqually(count, maxValue int) []int {
	input := make([]int, 0, count)
	n := 1

	for range count {
		input = append(input, n)
		n++
		if n > maxValue {
			n = 1
		}
	}

	return input
}

func SeedByRatio(count, maxValue int, ratio float64) []int {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	input := make([]int, 0, count)

	mainPartSize := int(float64(count) * ratio)
	randomPartSize := count - mainPartSize

	mainPartMaxValue := int(float64(maxValue) * (float64(1) - ratio))

	// Seed almost whole range with minimum diversity
	for range mainPartSize {
		input = append(input, rnd.Intn(mainPartMaxValue-1)+1)
	}

	// Other - random
	for range randomPartSize {
		input = append(input, rnd.Intn(maxValue-1)+1)
	}

	return input
}

// TODO: optimize
// func RoundToPrecision(val, minValue, maxValue float64) float64 {
// 	epsilon := math.Abs(maxValue - minValue)
// 	if epsilon == 0 {
// 		return val
// 	}

// 	decimalPlaces := int(math.Ceil(-math.Log10(epsilon)))
// 	if decimalPlaces < 0 {
// 		decimalPlaces = 0
// 	}

// 	factor := math.Pow(10, float64(decimalPlaces))
// 	return math.Round(val*factor) / factor
// }

func ToASCII(str string) byte {
	val, _ := strconv.Atoi(str)
	return byte(val)
}

func IsNthBitSet(num byte, n int) bool {
	var mask byte = 1 << n
	return (num & mask) != 0
}

func BytesToInts(bytes []byte) []int {
	ints := make([]int, 0, len(bytes))
	for _, b := range bytes {
		ints = append(ints, int(b))
	}
	return ints
}

func IntsToBytes(ints []int) []byte {
	bytes := make([]byte, 0, len(ints))
	for _, i := range ints {
		bytes = append(bytes, byte(i))
	}
	return bytes
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func FindLastIndex[T comparable](source, target []T) int {
	if len(target) == 0 || len(source) < len(target) {
		return -1
	}

	lastTarget := len(target) - 1
	i, t := len(source)-1, lastTarget

	for i >= 0 && t >= 0 {
		if source[i] == target[t] {
			t--
		} else {
			if t < lastTarget {
				t = lastTarget
			}
		}
		i--
	}

	if t >= 0 {
		return -1
	}

	return i + 1
}
