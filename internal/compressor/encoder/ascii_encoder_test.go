package encoder

import (
	"bytes"
	"math/rand"
	"testing"
)

const MAX_VALUE = 16255

func TestIntOverflowError(t *testing.T) {
	var buf bytes.Buffer
	var encoder ASCII

	value := MAX_VALUE + 1

	err := encoder.EncodeInt(&buf, value)
	if err == nil {
		t.Fatalf("expected error, got '%v'", err)
	}
	if err.Error() != "too big number" {
		t.Fatalf("unexpected error '%v'", err)
	}
}

func TestIntMaxValueOK(t *testing.T) {
	testIntSuccess(t, 16255)
}

func TestIntZeroOK(t *testing.T) {
	testIntSuccess(t, 0)
}

func TestIntRandom(t *testing.T) {
	testIntSuccess(t, rand.Intn(MAX_VALUE))
}

func testIntSuccess(t *testing.T, value int) {
	var buf bytes.Buffer
	var encoder ASCII

	err := encoder.EncodeInt(&buf, value)
	if err != nil {
		t.Fatalf("unexpected error '%v'", err)
	}

	num, err := encoder.DecodeInt(&buf)
	if err != nil {
		t.Fatalf("unexpected decoding error '%v'", err)
	}

	if num != value {
		t.Fatalf("expected '%q', got '%q'", value, num)
	}
}
