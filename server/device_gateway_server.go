package server

import (
	"net"
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
)

const QueueLen int64 = 218

type DeviceGatewayServer struct {
	connManager *common.SafeMap
	devManager  *DevicePKManager
	privateKey  []byte
	publicKey   []byte
}

//////////////////////////////////////////////////////////////////
//
//////////////////////////////////////////////////////////////////
func NewDeviceGatewayServer(conn *common.SafeMap, dev *DevicePKManager) *DeviceGatewayServer {
	common.Assert((conn != nil) && (dev != nil), "check manager nil")
	return &DeviceGatewayServer{connManager: conn, devManager: dev}
}

// the device point gateway listen on 8384
func (this *DeviceGatewayServer) Start(host string) error {
	common.Assert((this.connManager != nil) && (this.devManager != nil), "check manager nil")
	addr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		log.Errorf("Resolve TCP Addr Error:err[%v]", err)
		return err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Errorf("Listen Error:err[%v]", err)
		return err
	}
	for {
		socket, err := ln.AcceptTCP()
		if err != nil {
			log.Errorf("Accept Error:err[%v]", err)
			continue
		}
		log.Infof("Dev Connection begin:addr[%s]", socket.RemoteAddr())
		go this.handle(socket, 1024)
	}
	return err
}

func (this *DeviceGatewayServer) handle(socket *net.TCPConn, maxLen int) {
	conn := NewConnection(socket, maxLen, this.devManager)
	if conn != nil {
		defer conn.Close()
	} else {
		defer socket.Close()
	}
	// shake hands with the dev connection
	deviceGid, err := conn.HandShake()
	if err != nil || deviceGid == nil {
		log.Warningf("Device handshake failed exit:addr[%s]", socket.RemoteAddr())
		return
	}
	socket.SetKeepAlive(true)
	socket.SetKeepAlivePeriod(time.Second * 15)
	err = this.connManager.Insert(deviceGid, conn)
	if err != nil {
		log.Warningf("Insert the Connection Failed:err[%v]", err)
		return
	}

	// loop for all request and response
	log.Infof("Device connect succ:addr[%s], domain[%s], device[%s:%s]",
		socket.RemoteAddr(), deviceGid.domain, deviceGid.subDomain, deviceGid.deviceId)
	conn.Loop()
	log.Infof("Device Handler exit:addr[%s], domain[%s], device[%s:%s]",
		socket.RemoteAddr(), deviceGid.domain, deviceGid.subDomain, deviceGid.deviceId)

	// remove the device from device connection manager
	_, find := this.connManager.Delete(deviceGid)
	if !find {
		log.Warningf("Delete Device Connection before close:gid[%s], addr[%s]",
			deviceGid.String(), socket.RemoteAddr())
	}
}
