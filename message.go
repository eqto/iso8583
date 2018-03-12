/**
* Created by Visual Studio Code.
* User: tuxer
* Created At: 2018-02-26 18:21:12
**/

package iso

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"gitlab.com/tuxer/go-json"
	"gitlab.com/tuxer/go-logger"
	"sort"
	"strconv"
	"time"
)

const (
	//LengthASCII ...
	LengthASCII = `ascii`
	//LengthDecimal ...
	LengthDecimal = `decimal`
)

//Message ...
type Message map[string]interface{}

//messageData ...
type messageData map[int]interface{}

var (
	alphabeticalBits = []byte{37, 39, 42, 43, 44, 48, 55, 61, 62}
	specialBits      = []byte{42, 43, 44, 48, 55, 61, 52}
	llBits           = []byte{2, 32, 34, 44}
	lllBits          = []byte{48, 55, 61, 62, 63}
	bitLengthMap     = map[int]int{
		3: 6, 4: 12, 7: 10, 8: 8, 11: 6, 12: 6, 13: 4, 18: 4, 37: 12, 39: 2, 42: 15, 43: 40, 49: 3, 70: 3,
	}
	bitFormatMap = map[int]string{
		7:  `0102150405`, //MMDDhhmmss
		12: `150405`,     //hhmmss
		13: `0102`,       //MMDD
	}

	bitmapMap = map[string][]byte{
		`0810`: []byte{7, 11, 39, 70},
	}

	defaultLengthType = LengthASCII
)

//Parse lengthType: LengthASCII | LengthDecimal, data: []byte
func Parse(lengthType string, data []byte) (*Message, int) {
	var buff *buffer
	totalLen := 0
	if lengthType == LengthASCII {
		if len(data) < 2 {
			return nil, 0
		}
		length := (int(data[0]) * 256) + int(data[1])
		if len(data)-2 < length {
			return nil, 0
		}
		totalLen += (2 + length)
		buff = newBuffer(data[2:])
	} else if lengthType == LengthDecimal {
		if len(data) < 4 {
			return nil, 0
		}
		length, _ := strconv.Atoi(string(data[:4]))
		if len(data)-4 < length {
			return nil, 0
		}
		buff = newBuffer(data[4:])
		totalLen += (4 + length)
	}
	m := Message{}
	mti := buff.read(4)
	m.SetMTI(string(mti))
	hexBitmap := buff.read(16)

	bitmap, _ := hex.DecodeString(string(hexBitmap))
	if bitmap[0]&(0x01<<7) > 0 {
		secondBitmap, _ := hex.DecodeString(string(buff.read(16)))
		bitmap = append(bitmap, secondBitmap...)
	}

	var index rune
	for _, val := range bitmap {
		for i := 7; i >= 0; i-- {
			index++
			if val&(0x01<<uint(i)) > 0 {
				var length int
				if bytes.ContainsRune(lllBits, index) {
					length, _ = strconv.Atoi(string(buff.read(3)))
				} else if bytes.ContainsRune(llBits, index) {
					length, _ = strconv.Atoi(string(buff.read(2)))
				} else if fixLength, ok := bitLengthMap[int(index)]; ok {
					length = fixLength
				}
				data := buff.read(length)
				m.setData(int(index), data)
			}
		}
	}

	return &m, totalLen
}

//SetLengthType ...
func (m Message) SetLengthType(lengthType string) *Message {
	m[`length_type`] = lengthType
	return &m
}

//GetLengthType ...
func (m Message) GetLengthType() string {
	if lengthType, ok := m[`length_type`].(string); ok {
		return lengthType
	}
	return LengthASCII
}

//SetMTI ...
func (m Message) SetMTI(mti string) *Message {
	m[`mti`] = mti
	return &m
}

//GetMTI ...
func (m Message) GetMTI() string {
	return m[`mti`].(string)
}

//SetString ...
func (m Message) SetString(bit int, value string) *Message {
	return m.setData(bit, value)
}

//SetNumeric ...
func (m Message) SetNumeric(bit int, value int) *Message {
	return m.setData(bit, value)
}

//SetTime ...
func (m Message) SetTime(bit int, value time.Time) *Message {
	return m.setData(bit, value)
}

func (m Message) setData(bit int, value interface{}) *Message {
	if data, ok := m[`data`].(messageData); ok {
		data[bit] = value
	} else {
		msgData := messageData{}
		msgData[bit] = value
		m[`data`] = msgData
	}
	return &m
}

//ToJSON ...
func (m Message) ToJSON() (jsReturn *json.Object) {
	jsReturn = &json.Object{}
	bitmap := make([]byte, 8)
	mapData, ok := m[`data`].(messageData)
	if !ok {
		return nil
	}

	var keys []int
	for key := range mapData {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for _, key := range keys {
		if key > 64 && len(bitmap) == 8 {
			bitmap = append(bitmap, make([]byte, 8)...)
			bitmap[0] |= 0x01 << 7
		}
		charPos := (key - 1) / 8
		bitmap[charPos] |= 0x01 << (8 - uint(key-(charPos*8)))

		runeKey := rune(key)

		padding := `0`
		isAlphabetically := bytes.ContainsRune(alphabeticalBits, runeKey)
		isSpecial := bytes.ContainsRune(specialBits, runeKey)

		if isAlphabetically || isSpecial {
			padding = ``
		}

		str := m.GetString(key)
		if bytes.ContainsRune(llBits, runeKey) {
			str = fmt.Sprintf(`%02d%s`, len(str), str)
		} else if bytes.ContainsRune(lllBits, runeKey) {
			str = fmt.Sprintf(`%03d%s`, len(str), str)
		} else if format, ok := bitFormatMap[key]; ok {
			if formatTime, ok := mapData[key].(time.Time); ok {
				str = formatTime.Format(format)
			}
		}

		if length, ok := bitLengthMap[key]; ok {
			str = fmt.Sprintf(`%`+padding+strconv.Itoa(length)+`s`, str)
		}
		jsReturn.Put(`data.`+strconv.Itoa(key), str)
	}

	jsReturn.Put(`mti`, m.GetMTI())
	jsReturn.Put(`bitmap`, hex.EncodeToString(bitmap))
	return
}

//ToJSONFormatted ...
func (m Message) ToJSONFormatted() []byte {
	return m.ToJSON().ToFormattedBytes()
}

//ToBytes ...
func (m Message) ToBytes() []byte {
	bitmap := make([]byte, 8)
	mapData, ok := m[`data`].(messageData)
	if !ok {
		return nil
	}

	dataBuff := buffer{}
	var keys []int
	for key := range mapData {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for _, key := range keys {
		if key > 64 && len(bitmap) == 8 {
			bitmap = append(bitmap, make([]byte, 8)...)
			bitmap[0] |= 0x01 << 7
		}
		charPos := (key - 1) / 8
		bitmap[charPos] |= 0x01 << (8 - uint(key-(charPos*8)))

		runeKey := rune(key)

		padding := `0`
		isAlphabetically := bytes.ContainsRune(alphabeticalBits, runeKey)
		isSpecial := bytes.ContainsRune(specialBits, runeKey)

		if isAlphabetically || isSpecial {
			padding = ``
		}

		str := m.GetString(key)
		if bytes.ContainsRune(llBits, runeKey) {
			str = fmt.Sprintf(`%02d%s`, len(str), str)
		} else if bytes.ContainsRune(lllBits, runeKey) {
			str = fmt.Sprintf(`%03d%s`, len(str), str)
		} else if format, ok := bitFormatMap[key]; ok {
			if formatTime, ok := mapData[key].(time.Time); ok {
				str = formatTime.Format(format)
			}
		}

		if length, ok := bitLengthMap[key]; ok {
			str = fmt.Sprintf(`%`+padding+strconv.Itoa(length)+`s`, str)
		}

		dataBuff.writeString(str)
	}
	data := dataBuff.bytes()

	header := buffer{}
	header.writeString(m.GetMTI())
	header.writeString(hex.EncodeToString(bitmap))

	msg := append(append(header.bytes(), data...))
	length := len(msg)

	switch m.GetLengthType() {
	case LengthASCII:
		byteLength := make([]byte, 2)
		byteLength[0] = byte(length / 256)
		byteLength[1] = byte(length - (int(byteLength[0]) * 256))
		data = append(byteLength, data...)
		return append(byteLength, msg...)
	case LengthDecimal:
		log.D(`decimal`)
		return append([]byte(fmt.Sprintf(`%04d`, length)), msg...)
	}
	return nil
}

//GetString ...
func (m Message) GetString(bit int) string {
	data, ok := m[`data`].(messageData)
	if !ok {
		return ``
	}
	switch val := data[bit].(type) {
	case int:
		return strconv.Itoa(val)
	case string:
		return val
	case []byte:
		return string(val)
	}
	return ``
}

//GetInt ...
func (m Message) GetInt(bit int) int {
	data, ok := m[`data`].(messageData)
	if !ok {
		return 0
	}
	switch val := data[bit].(type) {
	case int:
		return val
	case string:
		i, e := strconv.Atoi(val)
		if e != nil {
			return 0
		}
		return i
	case []byte:
		str := string(val)
		i, e := strconv.Atoi(str)
		if e != nil {
			return 0
		}
		return i
	}
	return 0
}
