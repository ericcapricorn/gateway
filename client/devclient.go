package main

import (
	"bytes"
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
	//conn, err := net.Dial("tcp", "101.251.106.4:8384")
	//conn, err := net.Dial("tcp", "127.0.0.1:8384")
	conn, err := net.Dial("tcp", "192.168.1.114:8384")
	if err != nil {
		fmt.Println("Connect Error:", err)
		return
	}
	defer conn.Close()
	body := make([]byte, 1024)
	var report, request, response, empty common.Message
	// handshake step 1
	randKey := common.GenerateRandomKey(common.ZC_HS_MSG_LEN)
	request.Header.MsgCode = common.ZC_CODE_HANDSHAKE_1
	request.Header.PayloadLen = common.ZC_HS_MSG_LEN + common.ZC_HS_DEVICE_ID_LEN
	deviceId := "zzzzzzzzzzzz"
	fmt.Println("Generate random and id:", randKey, []byte(deviceId))
	request.Payload = make([]byte, request.Header.PayloadLen)
	copy(request.Payload[0:common.ZC_HS_MSG_LEN], randKey)
	copy(request.Payload[common.ZC_HS_MSG_LEN:request.Header.PayloadLen], []byte(deviceId))
	var contex common.EncryptContex
	contex.PublicKey = []byte("80138512665003396643737838315916663972728479914654754587175091902061894104953")
	contex.EncryptType = common.ZC_SEC_TYPE_RSA
	err = common.Send(conn, contex, &request, common.DEV_WRITE_TIMEOUT)
	if err != nil {
		fmt.Println("HANDSHAKE_1 failed:", err)
		return
	}
	fmt.Println("Write Handshake_1 succ")
	// handshake step 2
	contex.PrivateKey = common.PrivateKey
	err = common.Receive(conn, contex, &response, common.DEV_READ_TIMEOUT)
	if err != nil {
		fmt.Println("HANDSHAKE_2 failed:", err)
		return
	}
	sessionKey := make([]byte, common.ZC_HS_SESSION_KEY_LEN)
	copy(sessionKey, response.Payload[common.ZC_HS_MSG_LEN:])
	fmt.Println("Receive Handshake_2 Session Key:", response.Payload[:common.ZC_HS_MSG_LEN], sessionKey)

	// handshake step 3
	contex.EncryptType = common.ZC_SEC_TYPE_AES
	contex.SessionKey = sessionKey
	copy(request.Payload, randKey)
	request.Payload = request.Payload[0:common.ZC_HS_MSG_LEN]
	request.Header.PayloadLen = common.ZC_HS_MSG_LEN
	if len(request.Payload) != int(request.Header.PayloadLen) {
		fmt.Println("check payload len failed", len(request.Payload), request.Header.PayloadLen)
		return
	}
	request.Header.MsgCode = common.ZC_CODE_HANDSHAKE_3
	err = common.Send(conn, contex, &request, common.DEV_WRITE_TIMEOUT)
	if err != nil {
		fmt.Println("HANDSHAKE_3 failed:", err)
		return
	}
	fmt.Println("Write Handshake_3 succ")

	// handshake step 4
	err = common.Receive(conn, contex, &response, common.DEV_READ_TIMEOUT)
	if err != nil {
		fmt.Println("HANDSHAKE_4 failed:", err)
		return
	} else if !bytes.Equal(response.Payload, randKey) {
		fmt.Println("check random key failed")
		return
	} else {
		fmt.Println("CLOUD SAY HELLO WORLD TO DEV")
	}

	for i := 0; i < 3600; i++ {
		// step 1. send report message
		report.Header.PayloadLen = 0
		report.Header.Version = 123
		report.Header.MsgCode = common.ZC_CODE_HEARTBEAT
		report.Header.MsgId = uint8(i + 101)
		report.Payload = body[:report.Header.PayloadLen]
		err = common.Send(conn, contex, &report, common.DEV_WRITE_TIMEOUT)
		if err != nil {
			fmt.Println("Write Report Error:", err)
			return
		}
		// step 2. read the request
		err = common.Receive(conn, contex, &request, common.DEV_READ_TIMEOUT)
		if err != nil {
			fmt.Println("Receive Response Error:", err)
			return
		}
		fmt.Println("Read Request succ:", request.Header.MsgId, request.Payload)
		// step 3. echo empty message
		empty.Header.MsgId = request.Header.MsgId
		empty.Header.Version = request.Header.Version
		empty.Header.MsgCode = common.ZC_CODE_EMPTY
		empty.Header.PayloadLen = 0
		err = common.Send(conn, contex, &empty, common.DEV_WRITE_TIMEOUT)
		if err != nil {
			fmt.Println("Write Empty Response failed:", err)
			return
		}
		// fmt.Println("Write Empty Response", request.Header.MsgId)
		// step 4. write echo ack
		request.Header.MsgCode = common.ZC_CODE_ACK
		err = common.Send(conn, contex, &request, common.DEV_WRITE_TIMEOUT)
		if err != nil {
			fmt.Println("Write Response failed:", err)
			return
		}
		fmt.Println("Write Response:", request.Header.MsgId, request.Payload)
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 1; i++ {
		go func() {
			for {
				loop()
				time.Sleep(time.Second)
			}
		}()
	}
	common.WaitKill()
}
