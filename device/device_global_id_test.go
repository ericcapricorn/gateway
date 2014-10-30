package server

import (
	"bytes"
	"testing"
)

// the local key setted
func SetKey(d *DeviceInfo) {
	key := []byte{1, 2, 3, 4}
	key = append(key, 5, 7, 8)
	d.SetPublicKey(string(key))
}

func TestDeviceInfo(t *testing.T) {
	d := NewDeviceInfo("domain", "submain", "testing_device", "publickey")
	if d.Gid().DeviceId() != "testing_device" {
		t.Error("check device id failed")
	}
	d.Gid().SetDeviceId("tst_device")
	if d.Gid().DeviceId() != "tst_device" {
		t.Error("check device id failed")
	}
	key := []byte{1, 2, 3, 4, 5, 7, 8}
	d.SetPublicKey(string(key))
	if !bytes.Equal([]byte(d.PublicKey()), key) {
		t.Error("check public key failed")
	}
	// modify the public key
	key2 := []byte{2, 3, 4, 5}
	d.SetPublicKey(string(key2))
	if !bytes.Equal([]byte(d.PublicKey()), key2) {
		t.Error("check public key failed")
	}
	//
	SetKey(d)
	if !bytes.Equal([]byte(d.PublicKey()), key) {
		t.Error("check public key failed")
	}
}
