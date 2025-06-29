package compress

import "math/big"

type frequencyMap map[int]int

type probability struct {
	low  *big.Float
	high *big.Float
}

type ProbabilityMap map[int]probability
