package iso8583

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type messageData map[int]interface{}

type Message struct {
	deviceHeader string
	mti          string
	bitmap       []byte
	data         messageData
	keys         []int
	bitLength    bitLength
}

func (m *Message) SetDeviceHeader(deviceHeader string) {
	m.deviceHeader = deviceHeader
}

func (m *Message) GetDeviceHeader() string {
	return m.deviceHeader
}

func (m *Message) SetMTI(mti string) *Message {
	m.mti = mti
	return m
}

func (m *Message) GetMTI() string {
	return m.mti
}

func (m *Message) Get(bit int) []byte {
	if m.data == nil {
		return nil
	}
	switch val := m.data[bit].(type) {
	case int:
		return []byte(strconv.Itoa(val))
	case string:
		return []byte(val)
	case []byte:
		return val
	}
	return nil
}

func (m *Message) SetString(bit int, value string) *Message {
	return m.setData(bit, value)
}

func (m *Message) GetString(bit int) string {
	if m.data == nil {
		return ``
	}
	if format, ok := timeBit[bit]; ok {
		if formatTime, ok := m.data[bit].(time.Time); ok {
			return formatTime.Format(format)
		}
		return strings.Repeat(`0`, len(format))
	}
	switch val := m.data[bit].(type) {
	case int:
		return strconv.Itoa(val)
	case string:
		return val
	case []byte:
		return string(val)
	}
	return ``
}

func (m *Message) SetNumeric(bit int, value int) *Message {
	return m.setData(bit, value)
}

func (m *Message) SetTime(bit int, value time.Time) *Message {
	return m.setData(bit, value)
}

func (m *Message) GetInt(bit int) int {
	if m.data == nil {
		return 0
	}
	switch val := m.data[bit].(type) {
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

func (m *Message) Has(bit int) bool {
	if m.data != nil {
		if _, ok := m.data[bit]; ok {
			return true
		}
	}
	return false
}

func (m *Message) Clone() *Message {
	msg := Message{
		deviceHeader: m.deviceHeader,
		mti:          m.mti,
		data:         messageData{},
		bitmap:       make([]byte, len(m.bitmap)),
		keys:         make([]int, len(m.keys)),
	}
	copy(msg.bitmap, m.bitmap)
	copy(msg.keys, m.keys)
	for key, val := range m.data {
		msg.data[key] = val
	}
	return &msg
}

func (m *Message) Bytes() []byte {
	if m.data == nil {
		return nil
	}

	buff := buffer{}
	for _, key := range m.keys {
		str := m.GetString(key)

		runeKey := rune(key)

		if format, ok := timeBit[key]; ok {
			if formatTime, ok := m.data[key].(time.Time); ok {
				str = formatTime.Format(format)
			} else {
				str = strings.Repeat(`0`, len(format))
			}
		} else if length, ok := m.bitLength.Get(key); ok {
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

func (m *Message) String() string {
	return string(m.Bytes())
}

func (m *Message) Dump() string {
	buff := &strings.Builder{}
	if header := m.GetDeviceHeader(); header != `` {
		fmt.Fprintf(buff, "Device Header: %s\n", header)
	}
	fmt.Fprintf(buff, "MTI: %s\n", m.GetMTI())
	fmt.Fprintf(buff, "Bitmap: %s\n", m.BitmapString())
	for _, key := range m.keys {
		fmt.Fprintf(buff, "%3d: |%s|\n", key, m.GetString(key))
	}
	return buff.String()
}

func (m *Message) Bitmap() []byte {
	if m.bitmap == nil {
		m.bitmap = make([]byte, 8)
	}
	return m.bitmap
}

func (m *Message) BitmapString() string {
	return strings.ToUpper(hex.EncodeToString(m.Bitmap()))
}

func (m *Message) Unmarshal(dest interface{}) error {
	typeOf := reflect.TypeOf(dest)
	if typeOf.Kind() != reflect.Ptr {
		return errors.New(`dest is not a pointer`)
	}
	typeOf = typeOf.Elem()

	valOf := reflect.ValueOf(dest)
	valOf = valOf.Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if bit := field.Tag.Get(`bit`); bit != `` {
			if intBit, e := strconv.Atoi(bit); e == nil {
				kind := field.Type.Kind()
				val := valOf.Field(i)
				switch kind {
				case reflect.Int:
					val.SetInt(int64(m.GetInt(intBit)))
				case reflect.String:
					val.SetString(m.GetString(intBit))
				}
			}
		}
	}
	return nil
}

func (m *Message) setKey(key int) {
	if len(m.keys) == 0 {
		m.keys = []int{key}
	} else {
		pos := sort.SearchInts(m.keys, key)
		if pos < len(m.keys) && key == m.keys[pos] {
			return
		}
		m.keys = append(m.keys, 0)
		copy(m.keys[pos+1:], m.keys[pos:])
		m.keys[pos] = key
	}
}

func (m *Message) setData(bit int, value interface{}) *Message {
	bitmap := m.Bitmap()
	if bit > 64 && len(bitmap) == 8 {
		bitmap = append(bitmap, make([]byte, 8)...)
		bitmap[0] |= 0x01 << 7
	}
	pos := (bit - 1) / 8
	bitmap[pos] |= 0x01 << (8 - uint(bit-(pos*8)))
	m.bitmap = bitmap

	m.setKey(bit)

	if m.data == nil {
		m.data = messageData{}
	}
	m.data[bit] = value

	return m
}

func Parse(data []byte) (msg *Message, err error) {
	return new(Parser).Parse(data)
}
