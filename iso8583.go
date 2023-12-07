package iso8583

import (
	"bytes"
	"strconv"
)

var (
	alphabeticalBits = []byte{37, 38, 39, 40, 44, 45, 46, 47, 48, 49, 50, 51, 55, 61, 62}

	//28, 29, 30, 31, 97 is amount with C/D prefix
	specialBits = []byte{28, 29, 30, 31, 34, 35, 41, 42, 43, 52, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 96, 97, 98, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128}

	llBits  = []byte{2, 32, 33, 34, 35, 44, 45, 99, 100, 101, 102, 103}
	lllBits = []byte{36, 46, 47, 48, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127}

	stdLengthMap = map[int]int{
		3: 6, 4: 12, 5: 12, 6: 12, 7: 10, 8: 8, 9: 8, 10: 8,
		11: 6, 12: 6, 13: 4, 14: 4, 15: 4, 16: 4, 17: 4, 18: 4, 19: 3, 20: 3,
		21: 3, 22: 3, 23: 3, 24: 3, 25: 2, 26: 2, 27: 1, 28: 8, 29: 8, 30: 8,
		31: 8, 37: 12, 38: 6, 39: 2, 40: 3,
		41: 8, 42: 15, 43: 40, 49: 3, 50: 3,
		51: 3,
		70: 3, 90: 42,
	}

	timeBit = map[int]string{
		7:  `0102150405`, //MMDDhhmmss
		12: `150405`,     //hhmmss
		13: `0102`,       //MMDD
		15: `0102`,       //MMDD
	}
)

type bitLength struct {
	lengthMap map[int]int
}

func (b *bitLength) Set(bit, length int) {
	if b.lengthMap == nil {
		b.lengthMap = make(map[int]int)
	}
	b.lengthMap[bit] = length
}

func (b *bitLength) Get(bit int) (int, bool) {
	if b.lengthMap == nil {
		b.lengthMap = make(map[int]int)
	}
	if length, ok := b.lengthMap[bit]; ok {
		return length, true
	}
	if length, ok := stdLengthMap[bit]; ok {
		return length, true
	}
	return 0, false
}

func from(length int, pad byte, justify byte, data []byte) []byte {
	if length := length - len(data); length > 0 {
		padding := bytes.Repeat([]byte{pad}, length)
		if justify == 0 {
			return append(data, padding...)
		} else if justify == 1 {
			return append(padding, data...)
		}
	}
	return data
}

// FromString ...
func FromString(length int, data string) []byte {
	return from(length, ' ', 0, []byte(data))
}

// FromNumeric ...
func FromNumeric(length int, data int) []byte {
	return from(length, 0, 1, []byte(strconv.Itoa(data)))
}
