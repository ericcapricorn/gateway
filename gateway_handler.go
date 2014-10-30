package main

import (
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-gateway/device"
	"zc-service-go"
)

type GatewayServiceHandler struct {
	device *common.SafeMap
}

func NewGatewayServiceHandler(device *common.SafeMap) *GatewayServiceHandler {
	if device != nil {
		return &GatewayServiceHandler{device: device}
	}
	return nil
}

func (this *GatewayServiceHandler) handleForwardRequest(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	Conn, find := this.device.Find(*server.NewDeviceGID(domain, subDomain, deviceId))
	if !find || Conn == nil {
		resp.SetErr(common.ErrDeviceForward.Error())
		log.Warningf("the device gateway not find:gid[%s:%s:%s]", domain, subDomain, deviceId)
		return
	}
	// convert the request to message by loading by so
	var request common.Message
	request.Header.Version = 1
	request.Header.MsgId = 1
	request.Header.MsgCode = 123
	payload, _ := req.GetPayload()
	request.Header.PayloadLen = uint16(len(payload))
	request.Payload = payload

	// receive the device response result
	result := make(chan common.Message, 32)
	defer close(result)
	err := Conn.(*server.Connection).SendRequest(&request, &result)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("forward message to dev failed:gid[%s:%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return
	}
	log.Infof("forward message to device succ:code[%d], gid[%s:%s:%s]", request.Header.MsgCode, domain, subDomain, deviceId)

	// wait reponse from the device then response to the appclient
	response, err := this.waitResponse(request.Header.MsgId, result, common.APP_TIMEOUT)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("wait the current response failed:gid[%s:%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return
	}
	log.Infof("forward device request succ:gid[%s:%s:%s]", domain, subDomain, deviceId)
	resp.PutObject("response", zc.ZObject{"body": response.Payload})
	resp.SetAck()
}

func (this *GatewayServiceHandler) waitResponse(requestId uint8, result chan common.Message, timeout int64) (*common.Message, error) {
	timeoutChannel := time.After(time.Duration(timeout))
	var response common.Message
	for {
		select {
		case response = <-result:
			log.Infof("receive the repsonse from device:mid[%d]", response.Header.MsgId)
			if response.Header.MsgId == 0 {
				log.Warningf("check the response header failed")
				return nil, common.ErrInvalidStatus
			}
			if response.Header.MsgId == requestId {
				return &response, nil
			} else {
				log.Warningf("response not equal with request:response[%d], request[%d]", response.Header.MsgId, requestId)
				continue
			}
		case <-timeoutChannel:
			log.Warningf("wait the reponse timeout:request[%d], timeout[%d]", requestId, timeout)
			return nil, common.ErrTimeout
		}
	}
}
