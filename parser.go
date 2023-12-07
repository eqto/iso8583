package iso8583

import (
	"bytes"
	"encoding/hex"
	"errors"
	"time"
)

type Parser struct {
	bitLength bitLength
}

func (p *Parser) SetBitLength(bit, length int) {
	p.bitLength.Set(bit, length)
}

func (p *Parser) GetBitLength(bit int) (int, bool) {
	return p.bitLength.Get(bit)
}

func (p *Parser) Parse(data []byte) (msg *Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg = nil
			err = errors.New(`invalid format`)
		}
	}()
	msg = &Message{}

	if bytes.HasPrefix(data, []byte(`ISO`)) { //buang prefix
		msg.SetDeviceHeader(string(data[:12]))
		data = data[12:]
	}

	buff := newBuffer(data)

	msg.SetMTI(buff.ReadString(4))

	bitmap, _ := hex.DecodeString(buff.ReadString(16))
	if bitmap[0]&(0x01<<7) > 0 {
		secondBitmap, _ := hex.DecodeString(buff.ReadString(16))
		bitmap = append(bitmap, secondBitmap...)
	}
	msg.bitmap = bitmap

	var index int
	for _, val := range bitmap {
		for i := 7; i >= 0; i-- {
			index++
			if val&(0x01<<uint(i)) > 0 {
				var length int
				if bytes.ContainsRune(lllBits, rune(index)) {
					length = buff.ReadInt(3)
				} else if bytes.ContainsRune(llBits, rune(index)) {
					length = buff.ReadInt(2)
				} else if fixLength, ok := p.bitLength.Get(index); ok {
					length = fixLength
				}
				data := buff.Read(length)
				if data != nil {
					if format, ok := timeBit[index]; ok {
						parsed, e := time.Parse(format, string(data))
						if e == nil {
							msg.setData(index, parsed)
						} else {
							msg.setData(index, ``)
						}
					} else {
						msg.setData(index, data)
					}
				}
			}
		}
	}

	return msg, nil
}
