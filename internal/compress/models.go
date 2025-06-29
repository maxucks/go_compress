package compress

type frequencyMap map[int]int

type probability struct {
	pb   float64
	low  float64
	high float64
}

type ProbabilityMap map[int]probability
