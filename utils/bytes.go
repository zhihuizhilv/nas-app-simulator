package utils

func ReverseBytes(src []byte) []byte {
	srcTemp := make([]byte, len(src))
	copy(srcTemp[:], src)
	for i := len(srcTemp)/2 - 1; i >= 0; i-- {
		opp := len(srcTemp) - 1 - i
		srcTemp[i], srcTemp[opp] = srcTemp[opp], srcTemp[i]
	}
	return srcTemp
}
