package arithmetic

import "math/big"

type frequencyMap map[byte]int

type probability struct {
	low  *big.Float
	high *big.Float
}

type ProbabilityMap map[byte]probability
