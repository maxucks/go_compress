package utils

import (
	"os"
	"strconv"
)

type Frequency struct {
	Num       int
	Frequency int
}

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

func ReadLogfile() ([]byte, error) {
	data, err := os.ReadFile("logfile.txt")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SaveLogfile(data []byte) error {
	file, err := os.OpenFile("logfile.compressed", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
