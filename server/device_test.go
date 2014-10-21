package server 

import (
	"bytes"
	"testing"
)

// the local key setted
func SetKey(d *Device) {
	key := []byte{1, 2, 3, 4}
	key = append(key, 5, 7, 8)
	d.SetPublicKey(key)
}

func TestDevice(t *testing.T) {
	d := NewDevice("testing_device")
	if d.GetId() != "testing_device" {
		t.Error("check device id failed")
	}
	d.SetId("tst_device")
	if d.GetId() != "tst_device" {
		t.Error("check device id failed")
	}
	key := []byte{1, 2, 3, 4, 5, 7, 8}
	d.SetPublicKey(key)
	if !bytes.Equal(d.GetPublicKey(), key) {
		t.Error("check public key failed")
	}
	// modify the public key
	key2 := []byte{2, 3, 4, 5}
	d.SetPublicKey(key2)
	if !bytes.Equal(d.GetPublicKey(), key2) {
		t.Error("check public key failed")
	}
	//
	SetKey(d)
	if !bytes.Equal(d.GetPublicKey(), key) {
		t.Error("check public key failed")
	}
}
