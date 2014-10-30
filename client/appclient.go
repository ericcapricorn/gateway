package main

import (
	"fmt"
	"math/rand"
	"time"
	"zc-common-go/common"
)

func loop() {
	body := make([]byte, 1024)
	// TODO using http client request the device
	var request, response common.Message
	for i := 0; i < 3600*1000; i++ {
		request.Header.PayloadLen = uint16(rand.Intn(30) + 1)
		request.Header.Version = 123
		request.Header.MsgCode = common.ZC_CODE_ZDESCRIBE
		request.Header.MsgId = uint8(i + 101)
		var j uint16
		for j = 0; j < request.Header.PayloadLen; j++ {
			body[j] = byte(j + 1)
		}
		old := time.Now().UTC()
		request.Payload = body[:request.Header.PayloadLen]
		fmt.Println("Write Request:", request.Header, body[:request.Header.PayloadLen])
		// check the message
		if response.Header.MsgCode != common.ZC_CODE_ERR {
			fmt.Println("Read Response:", response.Header, body[:response.Header.PayloadLen])
			for j = 0; j < response.Header.PayloadLen; j++ {
				common.Assert(response.Payload[j] == body[j], "check the response failed")
			}
			fmt.Println("request timeused:", request.Header.MsgId, old, time.Now().UTC())
		} else {
			fmt.Println("Send Request to Device Failed")
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 1; i++ {
		go loop()
	}
	common.WaitKill()
}
