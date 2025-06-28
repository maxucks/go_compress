package compress

type frequencyMap map[int]int

type probability struct {
	pb   float64
	low  float64
	high float64
}

type ProbabilityMap map[int]probability

func (m ProbabilityMap) Find(value float64) (int, *probability) {
	for num, pb := range m {
		// var epsilon float64 = 1e-4
		// if value > pb.low-epsilon && value < pb.high+epsilon {
		if value < pb.high && value >= pb.low {
			return num, &pb
		}
	}
	return -1, nil
}
