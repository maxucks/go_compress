package compress

type frequencyMap map[int]int

type probability struct {
	pb   float64
	low  float64
	high float64
}

type ProbabilityMap map[int]probability

// Searches for the range containing the value and returns its data
func (m ProbabilityMap) MatchRange(value float64) (int, probability) {
	for num, pb := range m {
		if value < pb.high && value >= pb.low {
			return num, pb
		}
	}
	return -1, probability{}
}
