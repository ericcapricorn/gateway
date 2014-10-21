package server

import (
	"fmt"
	"net"
	"time"
	"zc-common-go/common"
)

const QueueLen int64 = 218

type DeviceServer struct {
	connManager *common.SafeMap
	devManager  *DeviceManager
	privateKey  []byte
	publicKey   []byte
}

//////////////////////////////////////////////////////////////////
//
//////////////////////////////////////////////////////////////////
func NewDeviceServer(conn *common.SafeMap, dev *DeviceManager) *DeviceServer {
	common.Assert((conn != nil) && (dev != nil), "check manager nil")
	return &DeviceServer{connManager: conn, devManager: dev}
}

// the device point gateway listen on 8384
func (this *DeviceServer) Start(host string) error {
	common.Assert((this.connManager != nil) && (this.devManager != nil), "check manager nil")
	addr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		fmt.Println("Resolve TCP Addr Error:", err)
		return err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Listen Error:", err)
		return err
	}
	for {
		socket, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		fmt.Println("Dev Connection begin:", socket.RemoteAddr())
		go this.handle(socket, 1024)
	}
	return err
}

func (this *DeviceServer) handle(socket *net.TCPConn, maxLen int) {
	// TODO handle shake protocol for auth
	device := NewConnection(socket, maxLen, this.devManager)
	if device != nil {
		defer device.Close()
	} else {
		defer socket.Close()
	}
	// shake hands with the dev connection
	key, err := device.HandShake()
	if err != nil {
		fmt.Println("Device handshake failed exit:", socket.RemoteAddr())
		return
	}
	socket.SetKeepAlive(true)
	socket.SetKeepAlivePeriod(time.Second * 10)
	err = this.connManager.Insert(key, device)
	if err != nil {
		fmt.Println("Insert the Connection Failed:", err)
		return
	}
	fmt.Println("Device handshake succ:", socket.RemoteAddr(), key)
	// start the connection Loop routine
	device.Loop()
	// WARNING TODO remove the device from the pool
	fmt.Println("Device Handler exit:", socket.RemoteAddr())
	// remove connection then close the connection
	_, find := this.connManager.Delete(key)
	if !find {
		fmt.Println("Delete Device Connection before close")
	}
}
