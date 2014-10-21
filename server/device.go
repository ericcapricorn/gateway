package server 

type Device struct {
	// device id
	deviceId string
	// device rsa public key
	publicKey []byte
}

func NewDevice(id string) *Device {
	return &Device{deviceId: id}
}

func (this *Device) GetId() string {
	return this.deviceId
}

func (this *Device) SetId(id string) {
	this.deviceId = id
}

func (this *Device) GetPublicKey() []byte {
	return this.publicKey
}

func (this *Device) SetPublicKey(key []byte) {
	this.publicKey = key
}
