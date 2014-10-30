package server

import (
	"zc-common-go/common"
)

// app ctrl request only with id and result queue
type DeviceCtrlRequest struct {
	// old request message id
	oldMessageId uint8
	// receive response result queue created by appserver
	responseQueue *chan common.Message
}

func NewDeviceCtrlRequest(id uint8, result *chan common.Message) *DeviceCtrlRequest {
	return &DeviceCtrlRequest{oldMessageId: id, responseQueue: result}
}
