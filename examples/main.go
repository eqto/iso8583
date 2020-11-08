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

	println(fmt.Sprintf(`Sign Request: %s`, req))
	resp, _ := iso8583.Parse(req.Bytes())
	println(fmt.Sprintf("== Parsed Sign Request ==\n%s", resp.Dump()))
	println(`== End Request ==`)

	// t.Log(msg.JSON().ToFormattedString())
	// t.Log(string(msg.Bytes()))

}
