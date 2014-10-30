package server

import (
	"net"
	"sync"
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
)

const QueueLen int = 218

type DeviceGatewayServer struct {
	// all device connection
	connManager *common.SafeMap
	// the device public key manager
	devManager *DevicePKManager
}

func NewDeviceGatewayServer(conn *common.SafeMap, dev *DevicePKManager) *DeviceGatewayServer {
	common.Assert((conn != nil) && (dev != nil), "check param nil")
	return &DeviceGatewayServer{connManager: conn, devManager: dev}
}

// the device point gateway start listen
func (this *DeviceGatewayServer) Start(host string) error {
	common.Assert((this.connManager != nil) && (this.devManager != nil), "check param nil")
	addr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		log.Errorf("resolve tcp addr error:err[%v]", err)
		return err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Errorf("listen error:err[%v]", err)
		return err
	}
	var waitGroup sync.WaitGroup
	for {
		socket, err := ln.AcceptTCP()
		if err != nil {
			log.Errorf("accept error:err[%v]", err)
			continue
		}
		log.Infof("device connect start:addr[%s]", socket.RemoteAddr())
		waitGroup.Add(1)
		go this.deviceRoutine(&waitGroup, socket, QueueLen)
	}
	waitGroup.Wait()
	return err
}

func (this *DeviceGatewayServer) deviceRoutine(waitGroup *sync.WaitGroup, socket *net.TCPConn, maxLen int) {
	defer waitGroup.Done()
	conn := NewConnection(socket, maxLen, this.devManager)
	defer conn.Close()
	// step 1. shake hands with the dev connection
	deviceGid, err := conn.DeviceHandShake()
	if err != nil {
		if deviceGid != nil {
			log.Errorf("device handshake failed exit:addr[%s], gid[%s]", socket.RemoteAddr(), deviceGid.String())
			return
		}
		log.Errorf("device handshake failed exit:addr[%s]", socket.RemoteAddr())
		return
	}

	// step 2. set the socket option as keep alive
	socket.SetKeepAlive(true)
	socket.SetKeepAlivePeriod(time.Second * 30)

	// step 3. record the connection for forward manager
	err = this.connManager.Insert(deviceGid, conn)
	if err != nil {
		log.Errorf("insert the device connection Failed:addr[%s], gid[%s], err[%v]",
			socket.RemoteAddr(), deviceGid.String(), err)
		return
	}
	// for debug info
	log.Infof("device connection created:addr[%s], gid[%s]", socket.RemoteAddr(), deviceGid.String())

	// step 4. loop forward all request and receive all response
	conn.Loop(waitGroup)

	// step 5. remove conn from device connection manager
	_, find := this.connManager.Delete(deviceGid)
	if !find {
		log.Errorf("delete device connection before close:addr[%s], gid[%s]", socket.RemoteAddr(), deviceGid.String())
	}
	// for debug info
	log.Infof("device connection closed:addr[%s], gid[%s]", socket.RemoteAddr(), deviceGid.String())
}
