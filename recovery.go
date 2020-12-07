package main

import (
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
)

func doGetLifeCycle(rw *saferw.SafeRW) {
	var req p2pprotocol.GetLifecycle
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.GET_LIFECYCLE, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.GET_LIFECYCLE_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.GET_LIFECYCLE_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.GetLifecycleResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal GetLifecycleResp fail, err:", err)
		return
	}

	loggermsg.Info("GetLifecycleResp:", resp)
}

func doGetRecoverProgress(rw *saferw.SafeRW) {
	var req p2pprotocol.GetRecoverProgress
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.GET_RECOVERYPROGRESS, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.GET_RECOVERYPROGRESS_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.GET_RECOVERYPROGRESS_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.GetRecoverProgressResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal GetRecoverProgressResp fail, err:", err)
		return
	}

	loggermsg.Info("GetRecoverProgressResp:", resp)
}