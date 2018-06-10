package goobj

func zigzagEncode(v int64) uint64 {
	return uint64(v<<1) ^ uint64(v>>63)
}

func zigzagDecode(v uint64) int64 {
	return int64(v<<63)>>63 ^ int64(v>>1)
}
