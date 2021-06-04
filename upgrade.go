package main

import (
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	. "gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"time"
)

func Upgrade(rw *SafeRW) {
	//doGetUpgrade(rw)
	doUpgrade(rw)
	go getUpgradeProgressLoop(rw)
}

func doGetUpgrade(rw *SafeRW) {
	loggermsg.Debug("doGetUpgrade begin")
	defer loggermsg.Debug("doGetUpgrade end")

	req := p2pprotocol.GetUpgrade {
		Nonce: rand.Uint32(),
	}

	sendMsg(p2pprotocol.GET_UPGRADE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.GET_UPGRADE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.GetUpgradeResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal GetUpgradeResp fail. err:", err)
		return
	}

	loggermsg.Info("GetUpgradeResp suc, resp:", resp)
}

func doUpgrade(rw *SafeRW) {
	loggermsg.Debug("doUpgrade begin")
	defer loggermsg.Debug("doUpgrade end")

	req := p2pprotocol.Upgrade {
		Nonce: rand.Uint32(),
	}

	sendMsg(p2pprotocol.DO_UPGRADE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.DO_UPGRADE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.UpgradeResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal UpgradeResp fail. err:", err)
		return
	}

	loggermsg.Info("UpgradeResp suc, resp:", resp)
}

func getUpgradeProgressLoop(rw *SafeRW) {
	for {
		progress := doGetUpgradeProgress(rw)
		loggermsg.Info("upgrade progress:", progress)
		if progress >= 100 {
			break
		}

		time.Sleep(time.Second)
	}
}

func doGetUpgradeProgress(rw *SafeRW) uint32 {
	loggermsg.Debug("doGetUpgradeProgress begin")
	defer loggermsg.Debug("doGetUpgradeProgress end")

	req := p2pprotocol.GetUpgradeProgress {
		Nonce: rand.Uint32(),
	}

	sendMsg(p2pprotocol.GET_UPGRADEPROGRESS, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return 0
	}

	if msgId != uint32(p2pprotocol.GET_UPGRADEPROGRESS_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return 0
	}

	var resp p2pprotocol.GetUpgradeProgressResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal GetUpgradeProgressResp fail. err:", err)
		return 0
	}

	loggermsg.Info("GetUpgradeProgressResp suc, resp:", resp)
	return resp.Percent
}