package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
	"zc-common-go/common"
)

//////////////////////////////////////////////////////////////////
//
//////////////////////////////////////////////////////////////////

func loop() {
	//conn, err := net.Dial("tcp", "101.251.106.4:9394")
	//conn, err := net.Dial("tcp", "192.168.1.114:9394")
	conn, err := net.Dial("tcp", "127.0.0.1:9394")
	if err != nil {
		fmt.Println("Connect Error:", err)
		return
	}
	defer conn.Close()
	body := make([]byte, 1024)
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
		err = common.Send(conn, common.NilContex, &request, common.DEV_WRITE_TIMEOUT)
		if err != nil {
			fmt.Println("Write Request Error:", err)
			return
		}
		fmt.Println("Write Request:", request.Header, body[:request.Header.PayloadLen])
		// read the same content
		err = common.Receive(conn, common.NilContex, &response, common.DEV_READ_TIMEOUT)
		if err != nil {
			fmt.Println("Receive Response Error:", err)
			return
		}
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
