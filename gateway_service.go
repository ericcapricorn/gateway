package main

import (
	"fmt"
	"runtime"
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-gateway/server"
	"zc-service-go"
)

type AppGatewayService struct {
	// app connection manager
	zc.ZService
	handler *GatewayServiceHandler
	// device manager find the device connection
	devManager *common.SafeMap
}

func NewAppGatewayService(dev *common.SafeMap, config *zc.ZServiceConfig) *AppGatewayService {
	handler := NewGatewayServiceHandler(dev)
	if handler == nil {
		return nil
	}
	service := &AppGatewayService{handler: handler}
	service.Init("zc-gateway", config)

	// forward request and wait the device response
	service.Handle("forward", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		handler.handleRequest(req, resp)
	}))
	return service
}

func Monitor(devs *common.SafeMap) {
	for {
		// not using lock for set
		fmt.Println("dev count:", devs.Len())
		time.Sleep(30 * time.Second)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	devConnManager := common.NewSafeMap()
	devManager := server.NewDevicePKManager("192.168.1.114:5354", "zc-dm")
	if devConnManager == nil || devManager == nil {
		log.Fatal("device connection manager and device manager failed")
		return
	}
	// device gateway server start
	deviceServer := server.NewDeviceGatewayServer(devConnManager, devManager)
	if deviceServer == nil {
		log.Fatal("new device gateway server failed")
		return
	}
	go deviceServer.Start("192.168.1.114:8384")
	go Monitor(devConnManager)

	// app gateway service start
	var serverConfig = &zc.ZServiceConfig{Port: "9394"}
	appService := NewAppGatewayService(devConnManager, serverConfig)
	if appService == nil {
		log.Fatal("new app gateway service failed")
		return
	}
	appService.Start()
}
