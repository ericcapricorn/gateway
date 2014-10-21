package main

import (
	"fmt"
	"runtime"
	"time"
	"zc-common-go/common"
	"zc-gateway/server"
)

func Monitor(devs *common.SafeMap, apps *common.SafeSet) {
	for {
		// not using lock for set
		fmt.Println("dev count:", devs.Len(), "app count:", apps.Len())
		time.Sleep(30 * time.Second)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	devConnManager := common.NewSafeMap()
	devManager := server.NewDeviceManager()
	DevServer := server.NewDeviceServer(devConnManager, devManager)
	// go DevServer.Start("101.251.106.4:8384")
	go DevServer.Start("192.168.1.114:8384")
	appConnManager := common.NewSafeSet()
	AppServer := server.NewAppServer(appConnManager, devConnManager)
	// go AppServer.Start("101.251.106.4:9394")
	go AppServer.Start("127.0.0.1:9394")

	go Monitor(devConnManager, appConnManager)
	common.WaitKill()
}
