package main

import (
	"fmt"

	"github.com/eqto/iso8583"
	"github.com/eqto/iso8583/examples/message"
)

func main() {
	//create 0800 request
	sign := &message.Signon{NetworkCode: 1}
	builder := iso8583.NewBuilder(`0800`, 7, 11, 48, 70)
	req, _ := builder.New(sign)

	data := req.Bytes()
	println(fmt.Sprintf(`Sign Request: %s`, string(data)))
	resp, _ := iso8583.Parse(data)
	println(resp)

	// t.Log(msg.JSON().ToFormattedString())
	// t.Log(string(msg.Bytes()))

}
