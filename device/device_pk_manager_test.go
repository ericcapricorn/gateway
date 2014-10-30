package server

import (
	"fmt"
	"testing"
	"zc-service-go"
)

const domain string = "domain"
const service string = "zc-dm"
const hostname string = "localhost:5354"

const count int = 30

func regist(t *testing.T) {
	client := zc.NewZServiceClient(hostname, service)
	for i := 0; i < count; i++ {
		subDomain := fmt.Sprintf("subdomain%d", i%3)
		deviceId := fmt.Sprintf("device%d", i+1)
		publicKey := fmt.Sprintf("key:%d", i*2)
		request := zc.NewZMsg()
		request.SetName("registdevice")
		request.PutString("domain", domain)
		request.PutString("submain", subDomain)
		request.PutString("deviceid", deviceId)
		request.PutString("publickey", publicKey)
		request.PutBool("master", true)
		response, err := client.Send(request)
		if err != nil {
			t.Error("regist device failed", err)
		}
		if response.IsErr() {
			t.Error("response regist device failed", response.GetErr())
		}
	}
}

func TestDevicePkManager(t *testing.T) {
	device := NewDevicePKManager(hostname, service)
	for i := 0; i < count; i++ {
		subDomain := fmt.Sprintf("subdomain%d", i%3)
		deviceId := fmt.Sprintf("device%d", i+1)
		device, err := device.Get(DeviceGID{domain, subDomain, deviceId})
		if err == nil && device == nil{
			t.Error("should not get the device key")
		}
	}

	// regist the devices
	regist(t)

	for i := 0; i < count; i++ {
		subDomain := fmt.Sprintf("subdomain%d", i%3)
		deviceId := fmt.Sprintf("device%d", i+1)
		publicKey := fmt.Sprintf("key:%d", i*2)
		// exist
		device, err := device.Get(DeviceGID{domain, subDomain, deviceId})
		if err != nil {
			t.Error("get the device key failed", err)
		} else if device.publicKey != publicKey {
			t.Error("check public key failed", device.publicKey, publicKey)
		}
	}

	// not exist
	subDomain := fmt.Sprintf("invalid")
	deviceId := fmt.Sprintf("invalid")
	_, err := device.Get(DeviceGID{domain, subDomain, deviceId})
	if err == nil {
		t.Error("should not get the device key")
	}

	// stop the server and it can access ok because of cache
	// TODO stop the service then re get
	for i := 0; i < count; i++ {
		subDomain := fmt.Sprintf("subdomain%d", i%3)
		deviceId := fmt.Sprintf("device%d", i+1)
		publicKey := fmt.Sprintf("key:%d", i*2)
		// find from cache
		device, err := device.Get(DeviceGID{domain, subDomain, deviceId})
		if err != nil {
			t.Error("get the device key failed not by cache hit", err)
		} else if device.publicKey != publicKey {
			t.Error("check public key failed", device.publicKey, publicKey)
		}
	}
}
