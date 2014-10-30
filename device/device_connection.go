package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"
	"zc-common-go/common"
	log "zc-common-go/glog"
)

// device connection
type Connection struct {
	// connection closed routine exit status
	exit bool
	// device global id for log info
	gid DeviceGID
	// read response message from socket
	socket net.Conn
	// dev request id maker
	lock         sync.Mutex
	devRequestId uint8
	// session key for every device
	sessionKey []byte
	// communication encryption contex
	contex common.EncryptContex
	// device manager
	deviceManager *DevicePKManager

	// mapping for app request + result channel
	requestMap *common.SafeMap

	// app request channel
	requestQueue chan common.Message
	// dev response empty channel
	emptyRespQueue chan common.Message
}

func NewConnection(socket net.Conn, maxLen int, devs *DevicePKManager) *Connection {
	return &Connection{
		exit:           false,
		socket:         socket,
		devRequestId:   0,
		deviceManager:  devs,
		requestMap:     common.NewSafeMap(),
		requestQueue:   make(chan common.Message, maxLen),
		emptyRespQueue: make(chan common.Message, maxLen),
	}
}

func (this *Connection) String() string {
	return this.socket.RemoteAddr().String() + this.gid.String()
}

func (this *Connection) Close() {
	this.exit = true
	this.socket.Close()
	close(this.requestQueue)
	close(this.emptyRespQueue)
}

func (this *Connection) getNewRequestId() uint8 {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.devRequestId++
	if this.devRequestId == 0 {
		this.devRequestId++
	}
	return this.devRequestId
}

// async send request to device and wait the response on result
func (this *Connection) SendRequest(packet *common.Message, result *chan common.Message) (err error) {
	common.Assert(packet != nil && result != nil, "check input param failed")
	oldId := packet.Header.MsgId
	log.Infof("receive new Request from app:msgId[%d]", oldId)
	request := NewDeviceCtrlRequest(oldId, result)
	// rewrite the msg id as the connection new id
	packet.Header.MsgId = this.getNewRequestId()
	err = this.requestMap.Insert(packet.Header.MsgId, request)
	if err != nil {
		log.Warningf("check request insert failed:old[%d], new[%d]", oldId, packet.Header.MsgId)
		return err
	}
	log.Infof("insert the request mapping succ:mid[%d], code[%d], dest[%v], gid[%s]",
		packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr(), this.gid.String())
	// if the dev connection closed panic will occured
	defer func() {
		packet.Header.MsgId = oldId
		info := recover()
		if info != nil {
			log.Warningf("the request queue is closed:err[%v]", info)
			err = common.ErrDeviceConnClosed
		}
	}()
	this.requestQueue <- *packet
	return nil
}

// WARNING can not be blocked by the response queue
func (this *Connection) handleDevMessage(packet *common.Message) {
	switch packet.Header.MsgCode {
	// handle device heartbeat only ignore it
	case common.ZC_CODE_HEARTBEAT:
		log.Infof("receive dev heartbeat:mid[%d]", packet.Header.MsgId)
	case common.ZC_CODE_ERR:
		log.Warningf("receive dev error:mid[%d], code[%d], dest[%s]", packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr())
		this.handleDevResponse(packet)
	// handle the empty response
	case common.ZC_CODE_EMPTY:
		this.emptyRespQueue <- *packet
	// handle other user defined message response
	default:
		if packet.Header.MsgCode > 100 {
			this.handleDevReportMessage(packet)
		} else {
			fmt.Println("receive dev ack:", packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr())
			this.handleDevResponse(packet)
		}
	}
}

// handle device response message forward to the result channel
func (this *Connection) handleDevResponse(packet *common.Message) {
	request, find := this.requestMap.Find(packet.Header.MsgId)
	if find {
		// reset to the old message id
		packet.Header.MsgId = request.(*DeviceCtrlRequest).oldMessageId
		defer func() {
			info := recover()
			if info != nil {
				log.Warningf("the response queue is closed:err[%v]", info)
			}
			_, exist := this.requestMap.Delete(packet.Header.MsgId)
			if !exist {
				log.Errorf("delete request failed:mid[%d], code[%d], dest[%s], gid[%s]",
					packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr(), this.gid.String())
			}
			log.Infof("delete the request mapping succ:mid[%d], code[%d], dest[%s], gid[%s]",
				packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr(), this.gid.String())
		}()
		// if closed will panic
		*(request.(*DeviceCtrlRequest).responseQueue) <- *packet
	} else {
		log.Errorf("check the dev response not find request:mid[%d], code[%d], dest[%s], gid[%s]",
			packet.Header.MsgId, packet.Header.MsgCode, this.socket.RemoteAddr(), this.gid.String())
	}
}

// handle device report message
func (this *Connection) handleDevReportMessage(Request *common.Message) {
	fmt.Println("[DEV] handle dev report message", this.String(), Request.Header)
}

// Loop wait dev empty reponse for the current request
// WARNING: ALL request to device will be queueed or blocked
// if not receive the empty reponse packet
func (this *Connection) waitEmptyResponse(request *common.Message) {
	for !this.exit {
		var emptyPacket common.Message
		select {
		case emptyPacket = <-this.emptyRespQueue:
			// fmt.Println("find a empty response", emptyPacket.Header)
			// must wait the same empty message id
			if emptyPacket.Header.MsgId == request.Header.MsgId {
				return
			} else {
				log.Errorf("empty msg id not the same as last request:mid[%d], requestId[%d]",
					emptyPacket.Header.MsgId, request.Header.MsgId)
			}
		}
	}
}

func (this *Connection) Loop(waitGroup *sync.WaitGroup) {
	// AFTER THIS ALL THE MESSAGE USING AES
	this.contex.EncryptType = common.ZC_SEC_TYPE_AES
	this.contex.SessionKey = this.sessionKey
	// process write queue in the new routine
	waitGroup.Add(1)
	go func() {
		// request packet
		var requestPacket common.Message
		for !this.exit {
			select {
			case requestPacket = <-this.requestQueue:
				// write the packet to read queue
				fmt.Println("[DEV] send the new Request Message to device:", requestPacket.Header.MsgId)
				err := common.Send(this.socket, this.contex, &requestPacket, common.DEV_WRITE_TIMEOUT)
				if err != nil {
					log.Warningf("forward the packet failed:err[%v]", err)
					this.exit = true
					break
				}
				// wait the empty response for the requestPacket
				// if not wait empty succ, can not send the next request
				// from the requestQueue....
				this.waitEmptyResponse(&requestPacket)
				fmt.Println("[DEV] wait Empty Response succ:", requestPacket.Header.MsgId)
			}
		}
		log.Infof("forward device request routine exit:dest[%v]", this.socket.RemoteAddr())
		waitGroup.Done()
	}()

	// process the read empty or response or report message
	var devMessage common.Message
	for !this.exit {
		// read the packet and dispatch to the processor
		err := common.Receive(this.socket, this.contex, &devMessage, common.DEV_READ_TIMEOUT)
		if err != nil {
			log.Warningf("read the packet failed:dest[%v], err[%v]", this.socket.RemoteAddr(), err)
			this.exit = true
			break
		}
		fmt.Println("[DEV] Receive message from device succ", devMessage.Header.String())
		// TODO if too long not receive dev message, close it
		// Handle Dev Message three different type
		// repsonse Ack + response Empty + request Report
		// WARNING:can not blocked here...
		this.handleDevMessage(&devMessage)
		fmt.Println("[DEV] Handle dev message finish:", time.Now().UTC(), devMessage.Header, this.socket.RemoteAddr())
	}
	log.Infof("device routine exit succ:dest[%v], gid[%s]", this.socket.RemoteAddr(), this.gid.String())
}

func (this *Connection) DeviceHandShake() (*DeviceGID, error) {
	var handShake common.Message
	// HANDSHAKE_1 using cloud private key
	this.contex.EncryptType = common.ZC_SEC_TYPE_RSA
	this.contex.PrivateKey = common.PrivateKey
	err := common.Receive(this.socket, this.contex, &handShake, common.DEV_READ_TIMEOUT)
	if err != nil {
		log.Warningf("read handShake_1 failed:err[%v]", err)
		return nil, err
	} else if handShake.Header.MsgCode != common.ZC_CODE_HANDSHAKE_1 {
		log.Warningf("check message code failed:code[%d]", handShake.Header.MsgCode)
		return nil, common.ErrInvalidMsg
	} else if handShake.Header.PayloadLen != (common.ZC_HS_MSG_LEN + common.ZC_HS_DEVICE_ID_LEN) {
		log.Warningf("check handshake step 1 failed:len[%d]", handShake.Header.PayloadLen)
		return nil, common.ErrInvalidMsg
	}
	fmt.Println("[DEV] HANDSHAKE_1 SUCC")

	// get device public key for rsa
	devRandom := handShake.Payload[0:common.ZC_HS_MSG_LEN]
	deviceId := handShake.Payload[common.ZC_HS_MSG_LEN : common.ZC_HS_MSG_LEN+common.ZC_HS_DEVICE_ID_LEN]
	log.Infof("Receive Dev Random and ID:", devRandom, deviceId)
	this.gid = DeviceGID{domain: "app", subDomain: "test", deviceId: string(deviceId)}
	device, err := this.deviceManager.Get(this.gid)
	if err != nil {
		log.Errorf("the device not valid:addr[%v], gid[%s]", this.socket.RemoteAddr(), this.gid.String())
		return nil, err
	}

	// HANDSHAKE_2 using device public key
	this.contex.EncryptType = common.ZC_SEC_TYPE_RSA
	this.contex.PublicKey = []byte(device.PublicKey())
	this.sessionKey = common.GenerateRandomKey(common.ZC_HS_SESSION_KEY_LEN)
	log.Infof("New session key:key[%v], dest[%v], gid[%s]", this.sessionKey, this.socket.RemoteAddr(), this.gid.String())
	handShake.Header.MsgId = this.getNewRequestId()
	var response common.Message
	response.Header.MsgCode = common.ZC_CODE_HANDSHAKE_2
	response.Header.PayloadLen = common.ZC_HS_MSG_LEN + common.ZC_HS_SESSION_KEY_LEN
	response.Payload = make([]byte, response.Header.PayloadLen)
	copy(response.Payload[0:common.ZC_HS_MSG_LEN], devRandom)
	copy(response.Payload[common.ZC_HS_MSG_LEN:response.Header.PayloadLen], this.sessionKey)
	err = common.Send(this.socket, this.contex, &response, common.DEV_WRITE_TIMEOUT)
	if err != nil {
		log.Warningf("Write ZC_CODE_HANDSHAKE_2 failed:err[%v]", err)
		return &this.gid, err
	}
	fmt.Println("[DEV] HANDSHAKE_2 SUCC")

	// HANDSHAKE_3 using session key
	this.contex.EncryptType = common.ZC_SEC_TYPE_AES
	this.contex.SessionKey = this.sessionKey
	err = common.Receive(this.socket, this.contex, &handShake, common.DEV_READ_TIMEOUT)
	if err != nil {
		log.Warningf("read handShake_3 failed:err[%v]", err)
		return &this.gid, err
	} else if handShake.Header.MsgCode != common.ZC_CODE_HANDSHAKE_3 {
		log.Warningf("check message code failed:code[%d]", handShake.Header.MsgCode)
		return &this.gid, common.ErrInvalidMsg
	} else if handShake.Header.PayloadLen != common.ZC_HS_MSG_LEN {
		log.Warningf("check handshake step 3 failed:len[%d]", handShake.Header.PayloadLen)
		return &this.gid, common.ErrInvalidMsg
	} else if !bytes.Equal(handShake.Payload, devRandom) {
		log.Warningf("check handshake content failed:payload[%v], random[%v]", handShake.Payload, devRandom)
		return &this.gid, common.ErrInvalidMsg
	}
	fmt.Println("[DEV] HANDSHAKE_3 SUCC")

	// HANDSHAKE_4 using session key
	this.contex.EncryptType = common.ZC_SEC_TYPE_AES
	this.contex.SessionKey = this.sessionKey
	handShake.Header.MsgId = this.getNewRequestId()
	handShake.Header.MsgCode = common.ZC_CODE_HANDSHAKE_4
	handShake.Header.PayloadLen = common.ZC_HS_MSG_LEN
	handShake.Payload = devRandom
	err = common.Send(this.socket, this.contex, &handShake, common.DEV_WRITE_TIMEOUT)
	if err != nil {
		log.Warningf("Write ZC_CODE_HANDSHAKE_4 failed:err[%v]", err)
		return &this.gid, err
	}
	fmt.Println("[DEV] HANDSHAKE_4 PASS, WELCOME:", this.socket.RemoteAddr())
	return &this.gid, err
}
