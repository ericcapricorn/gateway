package server

import (
	"zc-common-go/common"
)

type DeviceManager struct {
	manager *common.SafeMap
	// storage service
}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{manager: common.NewSafeMap()}
}

func (this *DeviceManager) Count() int {
	return this.manager.Len()
}

func (this *DeviceManager) Load() error {
	return nil
}

func (this *DeviceManager) Put(dev *Device) error {
	return this.manager.Insert(dev.GetId(), dev)
}

func (this *DeviceManager) Get(devId string) *Device {
	device, find := this.manager.Find(devId)
	if find {
		return device.(*Device)
	}
	return nil
}
