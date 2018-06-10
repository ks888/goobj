package goobj

import (
	"testing"
)

func TestZigZagEncodingAndDecoding(t *testing.T) {
	for i, testData := range []int64{
		0, -1, 1, -2, 2, int64(^uint(0) >> 1), -int64(^uint(0)>>1) - 1,
	} {
		encoded := zigzagEncode(testData)
		decoded := zigzagDecode(encoded)
		if testData != decoded {
			t.Errorf("input[%d]: %d, encoded: %d, decoded: %d", i, testData, encoded, decoded)
		}
	}
}
