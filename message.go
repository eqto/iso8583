package iso8583

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eqto/go-json"
)

//messageData ...
type messageData map[int]interface{}

//Message ...
type Message map[string]interface{}

//SetDeviceHeader ...
func (m Message) SetDeviceHeader(header string) {
	m[`header`] = header
}

//GetDeviceHeader ...
func (m Message) GetDeviceHeader() string {
	if header, ok := m[`header`]; ok {
		return header.(string)
	}
	return ``
}

//SetMTI ...
func (m Message) SetMTI(mti string) *Message {
	m[`mti`] = fmt.Sprintf(`%04s`, mti)
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

//Get ...
func (m Message) Get(bit int) []byte {
	data, ok := m[`data`].(messageData)
	if !ok {
		return nil
	}
	switch val := data[bit].(type) {
	case int:
		return []byte(strconv.Itoa(val))
	case string:
		return []byte(val)
	case []byte:
		return val
	}
	return nil
}

//GetString ...
func (m Message) GetString(bit int) string {
	data, ok := m[`data`].(messageData)
	if !ok {
		return ``
	}
	if format, ok := timeBit[bit]; ok {
		if formatTime, ok := data[bit].(time.Time); ok {
			return formatTime.Format(format)
		}
		return strings.Repeat(`0`, len(format))
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

//SetNumeric ...
func (m Message) SetNumeric(bit int, value int) *Message {
	return m.setData(bit, value)
}

//SetTime ...
func (m Message) SetTime(bit int, value time.Time) *Message {
	return m.setData(bit, value)
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

//Has ...
func (m Message) Has(bit int) bool {
	if data, ok := m[`data`].(messageData); ok {
		if _, ok := data[bit]; ok {
			return true
		}
	}
	return false
}

//Clone ...
func (m Message) Clone() Message {
	newM := Message{}
	for key, val := range m {
		newM[key] = val
	}
	return newM
}

//JSON ...
func (m Message) JSON() json.Object {
	js := make(json.Object)
	data, ok := m[`data`].(messageData)
	if !ok {
		return js
	}

	var keys []int
	for key := range data {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for _, key := range keys {
		str := m.GetString(key)

		runeKey := rune(key)

		if _, ok := timeBit[key]; ok {
		} else if length, ok := bitLengthMap[key]; ok {
			isAlphabetically := bytes.ContainsRune(alphabeticalBits, runeKey)
			isSpecial := bytes.ContainsRune(specialBits, runeKey)
			if isAlphabetically || isSpecial {
				str = fmt.Sprintf(`%-`+strconv.Itoa(length)+`s`, str)
			} else {
				str = fmt.Sprintf(`%0`+strconv.Itoa(length)+`v`, m.GetInt(key))
			}
		}

		js.Put(strconv.Itoa(key), str)
	}
	js.Put(`bitmap`, m.BitmapString())
	return js
}

//Bytes ...
func (m Message) Bytes() []byte {
	data, ok := m[`data`].(messageData)
	if !ok {
		return nil
	}
	var keys []int
	for key := range data {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	buff := buffer{}
	for _, key := range keys {
		str := m.GetString(key)

		runeKey := rune(key)

		if format, ok := timeBit[key]; ok {
			if formatTime, ok := data[key].(time.Time); ok {
				str = formatTime.Format(format)
			} else {
				str = strings.Repeat(`0`, len(format))
			}
		} else if length, ok := bitLengthMap[key]; ok {
			isAlphabetically := bytes.ContainsRune(alphabeticalBits, runeKey)
			isSpecial := bytes.ContainsRune(specialBits, runeKey)
			if isAlphabetically || isSpecial {
				str = fmt.Sprintf(`%-`+strconv.Itoa(length)+`s`, str)
			} else {
				str = fmt.Sprintf(`%0`+strconv.Itoa(length)+`d`, m.GetInt(key))
			}
		} else if bytes.ContainsRune(llBits, runeKey) {
			str = fmt.Sprintf(`%02d%s`, len(str), str)
		} else if bytes.ContainsRune(lllBits, runeKey) {
			str = fmt.Sprintf(`%03d%s`, len(str), str)
		}
		buff.writeString(str)
	}
	header := buffer{}
	devHeader := m.GetDeviceHeader()
	if len(devHeader) > 0 {
		header.writeString(devHeader)
	}
	header.writeString(m.GetMTI())
	header.writeString(m.BitmapString())

	return append(header.bytes(), buff.bytes()...)
}

func (m Message) String() string {
	return string(m.Bytes())
}

//Dump ...
func (m Message) Dump() string {
	buff := &strings.Builder{}
	if header := m.GetDeviceHeader(); header != `` {
		fmt.Fprintf(buff, "Device Header: %s\n", header)
	}
	fmt.Fprintf(buff, "MTI: %s\n", m.GetMTI())
	fmt.Fprintf(buff, "Bitmap: %s\n", m.BitmapString())
	return buff.String()
}

//Bitmap ...
func (m Message) Bitmap() []byte {
	if _, ok := m[`bitmap`]; !ok {
		m[`bitmap`] = make([]byte, 8)
	}
	return m[`bitmap`].([]byte)
}

//BitmapString ...
func (m Message) BitmapString() string {
	return strings.ToUpper(hex.EncodeToString(m.Bitmap()))
}

func (m Message) setData(bit int, value interface{}) *Message {
	bitmap := m.Bitmap()
	if bit > 64 && len(bitmap) == 8 {
		bitmap = append(bitmap, make([]byte, 8)...)
		bitmap[0] |= 0x01 << 7
	}
	pos := (bit - 1) / 8
	bitmap[pos] |= 0x01 << (8 - uint(bit-(pos*8)))
	m[`bitmap`] = bitmap

	if data, ok := m[`data`].(messageData); ok {
		data[bit] = value
	} else {
		msgData := messageData{}
		msgData[bit] = value
		m[`data`] = msgData
	}
	return &m
}

//Parse ...
func Parse(data []byte) (msg Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg = nil
			err = errors.New(`invalid format`)
		}
	}()
	msg = Message{}

	if bytes.HasPrefix(data, []byte(`ISO`)) { //buang prefix
		msg.SetDeviceHeader(string(data[:12]))
		data = data[12:]
	}

	buff := NewBuffer(data)

	msg.SetMTI(buff.ReadString(4))

	bitmap, _ := hex.DecodeString(buff.ReadString(16))
	if bitmap[0]&(0x01<<7) > 0 {
		secondBitmap, _ := hex.DecodeString(buff.ReadString(16))
		bitmap = append(bitmap, secondBitmap...)
	}

	var index rune
	for _, val := range bitmap {
		for i := 7; i >= 0; i-- {
			index++
			if val&(0x01<<uint(i)) > 0 {
				var length int
				if bytes.ContainsRune(lllBits, index) {
					length = buff.ReadInt(3)
				} else if bytes.ContainsRune(llBits, index) {
					length = buff.ReadInt(2)
				} else if fixLength, ok := bitLengthMap[int(index)]; ok {
					length = fixLength
				}
				data := buff.Read(length)
				if format, ok := timeBit[int(index)]; ok {
					parsed, e := time.Parse(format, string(data))
					if e == nil {
						msg.setData(int(index), parsed)
					} else {
						msg.setData(int(index), ``)
					}
				} else {
					msg.setData(int(index), data)
				}
			}
		}
	}

	return msg, nil
}
