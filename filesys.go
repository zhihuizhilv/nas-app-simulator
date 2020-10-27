package main

import (
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	. "gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
)

func doExplorDir(rw *SafeRW, dir string) {
	var req p2pprotocol.ExplorDir
	req.Nonce = rand.Uint32()
	req.Path = dir

	sendMsg(p2pprotocol.EXPLOR_DIR, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.EXPLOR_DIR_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.ExplorDirResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal ExplorDirResp fail. err:", err)
		return
	}

	loggermsg.Info("explor dir resp suc, dir:", dir, ", resp:", resp)
}

func doCreateDir(rw *SafeRW, dir string) {
	var req p2pprotocol.CreateDir
	req.Nonce = rand.Uint32()
	req.Path = dir

	sendMsg(p2pprotocol.CREATE_DIR, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.CREATE_DIR_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.CreateDirResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal CreateDirResp fail. err:", err)
		return
	}

	loggermsg.Info("create dir resp suc, root:", dir, ", resp:", resp)
}

func doDeleteFiles(rw *SafeRW, paths []string) {
	var req p2pprotocol.DeletePaths
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(p2pprotocol.DELETE_FILES, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.DELETE_FILES_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.DeletePathsResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal DeleteFilesResp fail. err:", err)
		return
	}

	loggermsg.Info("delete files resp, resp:", resp)
}

func doRenamePath(rw *SafeRW, path, newName string) {
	var req p2pprotocol.RenamePath
	req.Nonce = rand.Uint32()
	req.Path = path
	req.Name = newName

	sendMsg(p2pprotocol.RENAME_PATH, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.RENAME_PATH_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.RenamePathResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal RenamePathResp fail. err:", err)
		return
	}

	loggermsg.Info("rename path resp, resp:", resp)
}

func doPutinRecycle(rw *SafeRW, paths []string) {
	var req p2pprotocol.PutInRecycle
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(p2pprotocol.PUTIN_RECYCLE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.PUTIN_RECYCLE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.PutInRecycleResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal PutInRecycleResp fail. err:", err)
		return
	}

	loggermsg.Info("putin recycle resp, resp:", resp)
}

func doDredgeOutRecycle(rw *SafeRW, paths []string) {
	var req p2pprotocol.DredgeOutRecycle
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(p2pprotocol.DREDGEOUT_RECYCLE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := p2pprotocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(p2pprotocol.DREDGEOUT_RECYCLE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp p2pprotocol.DredgeOutRecycleResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal DredgeOutRecycleResp fail. err:", err)
		return
	}

	loggermsg.Info("dredge out recycle resp, resp:", resp)
}
