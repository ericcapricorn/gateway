package server

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-service-go"
)

type DevicePKManager struct {
	// device warehouse hostname and service name
	serviceHost string
	serviceName string
	// master device publickey container
	manager *common.SafeMap
}

func NewDevicePKManager(host, name string) *DevicePKManager {
	return &DevicePKManager{serviceHost: host, serviceName: name, manager: common.NewSafeMap()}
}

func (this *DevicePKManager) Count() int {
	return this.manager.Len()
}

func (this *DevicePKManager) Delete(id DeviceGID) {
	this.manager.Delete(id)
}

func (this *DevicePKManager) Put(dev *DeviceInfo) error {
	return this.manager.Insert(dev.gid, dev.publicKey)
}

func (this *DevicePKManager) Get(id DeviceGID) *DeviceInfo {
	publicKey, find := this.manager.Find(id)
	if find {
		return NewDeviceInfo(id.domain, id.subDomain, id.deviceId, publicKey.(string))
	} else {
		request := zc.NewZMsg()
		request.PutString("domain", id.domain)
		request.PutString("submain", id.subDomain)
		request.PutString("deviceid", id.deviceId)
		client := zc.NewZServiceClient(this.serviceHost, this.serviceName)
		response, err := client.Send(request)
		if err != nil {
			log.Warning("get device public key failed:domain[%s], device[%s:%s], err[%v]",
				id.domain, id.subDomain, id.deviceId, err)
			return nil
		}
		publicKey := response.GetString("publicKey")
		if len(publicKey) > 0 {
			dev := NewDeviceInfo(id.domain, id.subDomain, id.deviceId, publicKey)
			this.Put(dev)
			return dev
		} else {
			log.Errorf("master device public key invalid:domain[%s], device[%s:%s]",
				id.domain, id.subDomain, id.deviceId)
		}
	}
	return nil
}
