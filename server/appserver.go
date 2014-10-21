package server

import (
	"fmt"
	"net"
	"zc-common-go/common"
)

type AppServer struct {
	// app connection manager
	connManager *common.SafeSet
	// device manager find the device connection
	devManager *common.SafeMap
}

//////////////////////////////////////////////////////////////////
//
//////////////////////////////////////////////////////////////////
func NewAppServer(conn *common.SafeSet, dev *common.SafeMap) *AppServer {
	common.Assert(conn != nil, "check connection manager is nil")
	return &AppServer{connManager: conn, devManager: dev}
}

// the control point gateway listen on 9394
func (this *AppServer) Start(host string) (err error) {
	common.Assert(this.connManager != nil, "connection manager is nil")
	ln, err := net.Listen("tcp", host)
	if err != nil {
		fmt.Println("Listen Error:", err)
		return err
	}
	for {
		socket, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		fmt.Println("App Connection begin:", socket.RemoteAddr())
		go this.handle(socket)
	}
}

// ignore the not equal response id
func (this *AppServer) waitResponse(request *common.Message, socket net.Conn, result chan common.Message) error {
	var response common.Message
	for {
		fmt.Println("[APP] wait the response from device")
		select {
		case response = <-result:
			fmt.Println("[APP] receive the repsonse from device:", response.Header.MsgId)
			if response.Header.MsgId == request.Header.MsgId {
				err := common.Send(socket, common.NilContex, &response, common.APP_TIMEOUT)
				if err != nil {
					fmt.Println("write response failed:", err)
					return err
				}
				return nil
			} else {
				fmt.Println("check response not equal with cur request:",
					response.Header.MsgId, request.Header.MsgId)
				continue
			}
			// case timeout
		}
	}
	return nil
}

func (this *AppServer) handle(socket net.Conn) {
	defer socket.Close()
	this.connManager.Add(socket.RemoteAddr())
	var errorResponse common.Message
	errorResponse.Header.MsgCode = common.ZC_CODE_ERR
	errorResponse.Header.PayloadLen = 0
	var request common.Message
	result := make(chan common.Message, 1024)
	for {
		fmt.Println("[APP] wait receive the next APP request")
		// read the packet and dispatch to the processor
		err := common.Receive(socket, common.NilContex, &request, common.APP_TIMEOUT)
		if err != nil {
			fmt.Println("read packet failed:", err)
			break
		}
		fmt.Println("[APP] receive app message:", request.Header.MsgId, socket.RemoteAddr())
		// WARNING TODO find the device id then put the packet to the request queue
		deviceId := "zzzzzzzzzzzz"
		devConn, find := this.devManager.Find(deviceId)
		if !find {
			fmt.Println("the device not online rightnow:", deviceId, socket.RemoteAddr())
			errorResponse.Header.MsgId = request.Header.MsgId
			err := common.Send(socket, common.NilContex, &errorResponse, common.APP_TIMEOUT)
			if err != nil {
				fmt.Println("write error response to app failed:", err)
				break
			}
			continue
		}

		// forward the request to the device
		err = devConn.(*Connection).Write(&request, &result)
		if err != nil {
			fmt.Println("forward message to dev failed:", deviceId, socket.RemoteAddr())
			errorResponse.Header.MsgId = request.Header.MsgId
			err := common.Send(socket, common.NilContex, &errorResponse, common.APP_TIMEOUT)
			if err != nil {
				fmt.Println("write error response to app failed:", err)
				break
			}
			continue
		}
		// wait reponse from the device then response to the appclient
		err = this.waitResponse(&request, socket, result)
		if err != nil {
			fmt.Println("wait the current response failed:", deviceId, socket.RemoteAddr(), err)
			break
		}
		fmt.Println("[APP] wait the response from device:", deviceId, socket.RemoteAddr())
	}
	close(result)
	this.connManager.Remove(socket.RemoteAddr())
	// WARNING TODO remove the device from the pool
	fmt.Println("[App] connection exit:", socket.RemoteAddr())
}
