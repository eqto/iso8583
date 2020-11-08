package iso8583

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

//Builder ...
type Builder struct {
	mti  string
	bits []byte
}

//New ...
func (m *Builder) New(data interface{}) (*Message, error) {
	elem := reflect.TypeOf(data)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return nil, errors.New(`invalid parameter data`)
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	msg := Message{}
	msg.SetMTI(m.mti)
	len := elem.NumField()
	for i := 0; i < len; i++ {
		tag := elem.Field(i).Tag
		if bitTag := tag.Get(`bit`); bitTag != `` {
			if bit, e := strconv.Atoi(bitTag); e == nil {
				if bytes.ContainsRune(m.bits, rune(bit)) {
					switch kind := val.Field(i).Kind(); kind {
					case reflect.Int:
						msg.SetNumeric(bit, int(val.Field(i).Int()))
					case reflect.String:
						msg.SetString(bit, val.Field(i).String())
					default:
						return nil, fmt.Errorf(`invalid data type of %s:%s, only support string and int`, elem.Field(i).Name, kind.String())
					}
				}
			} else {
				return nil, fmt.Errorf(`invalid tag for bit, please use int value. Current tag: %v`, tag)
			}
		}
	}
	//set mandatory fields that has not set
	time := time.Now()
	for _, bit := range m.bits {
		intBit := int(bit)
		if !msg.Has(intBit) {
			if _, ok := timeBit[intBit]; ok {
				msg.SetTime(intBit, time)
			} else {
				msg.SetString(intBit, ``)
			}
		}
	}

	return &msg, nil
}

//NewBuilder ...
func NewBuilder(mti string, bits ...byte) Builder {
	return Builder{mti: mti, bits: bits}
}
