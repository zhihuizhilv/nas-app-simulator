package main

import (
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"time"
)

func doReset(rw *saferw.SafeRW) {
	var req p2pprotocol.ResetFactorySetting
	req.Nonce = rand.Uint32()
	req.Timestamp = uint64(time.Now().Unix())
	req.Signature = nil

	sendMsg(p2pprotocol.RESET_FACTORYSETTING, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.RESET_FACTORYSETTING_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.ResetFactorySettingResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal ResetFactorySettingResp fail. err:", err)
		return
	}

	loggermsg.Info("reset factory resp, resp:", resp)
}
