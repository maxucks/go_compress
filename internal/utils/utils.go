package utils

import (
	"math"
	"math/rand"
	"strconv"
)

type Frequency struct {
	Num       int
	Frequency int
}

func SeedNumbers(count, maxValue int) []int {
	input := make([]int, 0, count)
	for range count {
		input = append(input, rand.Intn(maxValue-1)+1)
	}
	return input
}

func RoundToPrecision(val, minValue, maxValue float64) float64 {
	epsilon := math.Abs(maxValue - minValue)
	if epsilon == 0 {
		return val
	}

	decimalPlaces := int(math.Ceil(-math.Log10(epsilon)))
	if decimalPlaces < 0 {
		decimalPlaces = 0
	}

	factor := math.Pow(10, float64(decimalPlaces))
	return math.Round(val*factor) / factor
}

func ToASCII(str string) byte {
	val, _ := strconv.Atoi(str)
	return byte(val)
}

func IsNthBitSet(num byte, n int) bool {
	var mask byte = 1 << n
	return (num & mask) != 0
}
