package server

import (
	"fmt"
)

// Device Global Identifier
type DeviceGID struct {
	domain    string
	subDomain string
	deviceId  string
}

func NewDeviceGID(domain, subDomain, deviceId string) *DeviceGID {
	return &DeviceGID{domain: domain, subDomain: subDomain, deviceId: deviceId}
}

func (this *DeviceGID) String() string {
	return fmt.Sprintf("%s:%s:%s", this.domain, this.subDomain, this.deviceId)
}

func (this *DeviceGID) Domain() string {
	return this.domain
}

func (this *DeviceGID) SetDomain(domain string) {
	this.domain = domain
}

func (this *DeviceGID) SubDomain() string {
	return this.subDomain
}

func (this *DeviceGID) SetSubDomain(subDomain string) {
	this.subDomain = subDomain
}

func (this *DeviceGID) DeviceId() string {
	return this.deviceId
}

func (this *DeviceGID) SetDeviceId(id string) {
	this.deviceId = id
}

// device info
type DeviceInfo struct {
	gid       DeviceGID
	publicKey string
}

func NewDeviceInfo(domain, subDomain, deviceId, publicKey string) *DeviceInfo {
	return &DeviceInfo{gid: *NewDeviceGID(domain, subDomain, deviceId), publicKey: publicKey}
}

func (this *DeviceInfo) Gid() *DeviceGID {
	return &this.gid
}

func (this *DeviceInfo) SetGid(gid DeviceGID) {
	this.gid = gid
}

func (this *DeviceInfo) PublicKey() string {
	return this.publicKey
}

func (this *DeviceInfo) SetPublicKey(key string) {
	this.publicKey = key
}
