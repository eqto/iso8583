package iso8583

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Builder struct {
	allowedBits   []byte
	mandatoryBits []byte
}

func (b *Builder) SetMandatoryBits(bits ...byte) {
	for _, bit := range bits {
		if !bytes.ContainsRune(b.allowedBits, rune(bit)) {
			b.allowedBits = append(b.allowedBits, bit)
		}
	}
	b.mandatoryBits = bits
}

func (b *Builder) New(mti string, data interface{}) (*Message, error) {
	msg := Message{}
	msg.SetMTI(mti)
	if m, ok := data.(map[int]interface{}); ok {
		for key, val := range m {
			switch val := val.(type) {
			case string:
				msg.SetString(key, val)
			case int:
				msg.SetNumeric(key, val)
			case time.Time:
				msg.SetTime(key, val)
			case float64:
				msg.SetNumeric(key, int(val))
			}
		}
	} else {
		typeOf := reflect.TypeOf(data)
		if typeOf.Kind() == reflect.Ptr {
			typeOf = typeOf.Elem()
		}
		if typeOf.Kind() != reflect.Struct {
			return nil, errors.New(`invalid parameter data`)
		}
		valData := reflect.ValueOf(data)
		if valData.Kind() == reflect.Ptr {
			valData = valData.Elem()
		}
		numFields := typeOf.NumField()
		for i := 0; i < numFields; i++ {
			tag := typeOf.Field(i).Tag
			bitTag, ok := tag.Lookup(`bit`)
			if !ok {
				continue
			}
			omitEmpty := false
			split := strings.Split(bitTag, `,`)
			if len(split) == 2 && split[1] == `omitempty` {
				omitEmpty = true
			}
			bit, e := strconv.Atoi(split[0])
			if e != nil {
				continue
			}
			if !bytes.ContainsRune(b.allowedBits, rune(bit)) {
				continue
			}
			val := valData.Field(i)
			kind := val.Kind()
			if kind == reflect.Ptr {
				if val.IsNil() {
					if !omitEmpty {
						msg.SetString(bit, ``)
					}
					continue
				} else {
					val = val.Elem()
				}
			}
			switch kind {
			case reflect.Int:
				msg.SetNumeric(bit, int(val.Int()))
			case reflect.String:
				msg.SetString(bit, val.String())
			case reflect.Struct:
				if time, ok := val.Interface().(time.Time); ok {
					msg.SetTime(bit, time)
					continue
				}
				fallthrough
			default:
				return nil, fmt.Errorf(`invalid data type of %s:%s, only support string, int, and time.time`, typeOf.Field(i).Name, kind.String())
			}

		}
	}

	//set mandatory fields that has not set
	time := time.Now()
	for _, bit := range b.mandatoryBits {
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

func NewBuilder(allowedBits ...byte) *Builder {
	return &Builder{allowedBits: allowedBits}
}
