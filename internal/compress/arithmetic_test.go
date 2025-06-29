package compress

import (
	"bytes"
	"maps"
	"testing"
)

func TestMetaEncodeOK(t *testing.T) {
	var buf bytes.Buffer
	compressor := NewArithmeticCompressor()

	numsCount := 8
	frmap := frequencyMap{1: 1, 2: 2}

	if err := compressor.encodeMeta(&buf, numsCount, frmap); err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	decodedNumsCount, decodedFrmap, err := compressor.decodeMeta(&buf)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	if numsCount != decodedNumsCount {
		t.Errorf("numsCount '%v' not equal to decoded '%v'", numsCount, decodedNumsCount)
	}
	if !maps.Equal(frmap, decodedFrmap) {
		t.Errorf("frmap '%v' not equal to decoded '%v'", frmap, decodedFrmap)
	}
}
