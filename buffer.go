/**
* Created by Visual Studio Code.
* User: tuxer
* Created At: 2018-02-27 22:50:53
**/

package iso

import (
	"bytes"
)

//Buffer ...
type buffer struct	{
	byteBuff		*bytes.Buffer

}

func newBuffer(data []byte) *buffer	{
	buffer := &buffer{}
	buffer.byteBuff = bytes.NewBuffer(data)
	return buffer
}

func (b *buffer) read(length int) []byte	{
	if b.byteBuff == nil	{
		b.byteBuff = &bytes.Buffer{}
	}
	data := make([]byte, length)
	if _, e := b.byteBuff.Read(data); e != nil	{
		println(e.Error())
	}
	return data
}

func (b *buffer) writeString(data string) *buffer	{
	if b.byteBuff == nil	{
		b.byteBuff = &bytes.Buffer{}
	}
	b.byteBuff.WriteString(data)
	return b
}

func (b *buffer) write(data []byte) *buffer	{
	if b.byteBuff == nil	{
		b.byteBuff = &bytes.Buffer{}
	}
	b.byteBuff.Write(data)
	return b
}

func (b *buffer) bytes() []byte	{
	return b.byteBuff.Bytes()
}