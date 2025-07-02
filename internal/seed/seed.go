package seed

import (
	"fmt"
	"math/rand"
	"time"
)

type SeedFn func() []int

func SeedRandom(count, maxValue int) (string, SeedFn) {
	fn := func() []int {
		src := rand.NewSource(time.Now().UnixNano())
		rnd := rand.New(src)

		input := make([]int, 0, count)
		for range count {
			input = append(input, rnd.Intn(maxValue-1)+1)
		}

		return input
	}

	description := fmt.Sprintf("size = %v, with values from 1 to %v, distributed randomly", count, maxValue)
	return description, fn
}

func SeedEqually(mul, maxValue int) (string, SeedFn) {
	count := maxValue * mul

	fn := func() []int {
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

	description := fmt.Sprintf("size = %v, with values from 1 to %v, distributed equally", count, maxValue)
	return description, fn
}

func SeedByRatio(count, maxValue int, ratio float64) (string, SeedFn) {
	mainPartSize := int(float64(count) * ratio)
	randomPartSize := count - mainPartSize
	mainPartMaxValue := int(float64(maxValue) * (float64(1) - ratio))

	if mainPartMaxValue < 2 {
		mainPartMaxValue = 2
	}

	fn := func() []int {
		src := rand.NewSource(time.Now().UnixNano())
		rnd := rand.New(src)

		input := make([]int, 0, count)

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

	description := fmt.Sprintf("size = %v, with values from 1 to %v, where %v%% of values are from 1 to %v, others are random", count, maxValue, ratio*100, mainPartMaxValue)
	return description, fn
}
