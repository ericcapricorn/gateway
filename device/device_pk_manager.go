package server

import (
	"errors"
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-service-go"
)

type DevicePKManager struct {
	// device warehouse hostname and service name
	serviceHost string
	serviceName string
	// master device publickey cache
	cache *common.LRUCache
}

const MAX_DEVICE_COUNT int64 = 100000

func NewDevicePKManager(host, name string) *DevicePKManager {
	return &DevicePKManager{serviceHost: host, serviceName: name, cache: common.NewLRUCache(MAX_DEVICE_COUNT)}
}

func (this *DevicePKManager) Delete(id DeviceGID) {
	this.cache.Delete(id)
}

func (this *DevicePKManager) Put(dev *DeviceInfo) {
	this.cache.Set(dev.gid, dev.publicKey)
}

func (this *DevicePKManager) Get(id DeviceGID) (*DeviceInfo, error) {
	publicKey, find := this.cache.Get(id)
	if find {
		dev := NewDeviceInfo(id.domain, id.subDomain, id.deviceId, publicKey.(string))
		log.Infof("get device public key from cache succ:domain[%s], device[%s:%s], key[%s]",
			id.domain, id.subDomain, id.deviceId, publicKey.(string))
		return dev, nil
	} else {
		request := zc.NewZMsg()
		request.SetName("getpublickey")
		request.PutString("domain", id.domain)
		request.PutString("submain", id.subDomain)
		request.PutString("deviceid", id.deviceId)
		client := zc.NewZServiceClient(this.serviceHost, this.serviceName)
		response, err := client.Send(request)
		if err != nil {
			log.Warningf("get device public key failed:domain[%s], device[%s:%s], err[%v]",
				id.domain, id.subDomain, id.deviceId, err)
			return nil, err
		}
		if response.IsErr() {
			log.Warningf("get device public key failed:domain[%s], device[%s:%s], err[%s]",
				id.domain, id.subDomain, id.deviceId, response.GetErr())
			return nil, errors.New(response.GetErr())
		}
		publicKey := response.GetString("publickey")
		if len(publicKey) > 0 {
			dev := NewDeviceInfo(id.domain, id.subDomain, id.deviceId, publicKey)
			this.Put(dev)
			return dev, nil
		} else {
			log.Errorf("master device public key invalid:domain[%s], device[%s:%s]",
				id.domain, id.subDomain, id.deviceId)
			return nil, common.ErrInvalidDevice
		}
	}
}
