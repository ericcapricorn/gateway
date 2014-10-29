package main

import (
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-gateway/server"
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

func (this *GatewayServiceHandler) waitResponse(requestId uint8, result chan common.Message, timeout int64) (*common.Message, error) {
	timeoutChannel := time.After(time.Duration(timeout))
	var response common.Message
	for {
		select {
		case response = <-result:
			log.Infof("receive the repsonse from device:id[%d]", response.Header.MsgId)
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

func (this *GatewayServiceHandler) handleRequest(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	Conn, find := this.device.Find(*server.NewDeviceGID(domain, subDomain, deviceId))
	if !find || Conn == nil {
		resp.SetErr(common.ErrDeviceOffline.Error())
		log.Warningf("the device not online:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return
	}

	// convert the request to message by loading by so
	var request common.Message

	result := make(chan common.Message, 64)
	defer close(result)
	// forward the message to the device
	err := Conn.(*server.Connection).Write(&request, &result)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("forward message to dev failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return
	}

	// wait reponse from the device then response to the appclient
	response, err := this.waitResponse(request.Header.MsgId, result, common.APP_TIMEOUT)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("wait the current response failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return
	}
	log.Infof("forward device request succ:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
	resp.PutObject("response", zc.ZObject{"body": response.Payload})
	resp.SetAck()
}
