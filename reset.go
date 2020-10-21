package main

import (
	"gitlab.dabank.io/nas/go-nas/p2p/protocol"
	"gitlab.dabank.io/nas/go-nas/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"time"
)

func doReset(rw *saferw.SafeRW) {
	var req protocol.ResetFactorySetting
	req.Nonce = rand.Uint32()
	req.Timestamp = uint64(time.Now().Unix())
	req.Signature = nil

	sendMsg(protocol.RESET_FACTORYSETTING, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.RESET_FACTORYSETTING_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.ResetFactorySettingResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal ResetFactorySettingResp fail. err:", err)
		return
	}

	loggermsg.Info("reset factory resp, resp:", resp)
}
