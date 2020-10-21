package main

import (
	"gitlab.dabank.io/nas/go-nas/p2p/protocol"
	"gitlab.dabank.io/nas/go-nas/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
)

func doExplorDir(rw *saferw.SafeRW, dir string) {
	var req protocol.ExplorDir
	req.Nonce = rand.Uint32()
	req.Path = dir

	sendMsg(protocol.EXPLOR_DIR, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.EXPLOR_DIR_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.ExplorDirResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal ExplorDirResp fail. err:", err)
		return
	}

	loggermsg.Info("explor dir resp suc, dir:", dir, ", resp:", resp)
}

func doCreateDir(rw *saferw.SafeRW, dir string) {
	var req protocol.CreateDir
	req.Nonce = rand.Uint32()
	req.Path = dir

	sendMsg(protocol.CREATE_DIR, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.CREATE_DIR_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.CreateDirResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal CreateDirResp fail. err:", err)
		return
	}

	loggermsg.Info("create dir resp suc, root:", dir, ", resp:", resp)
}

func doDeleteFiles(rw *saferw.SafeRW, paths []string) {
	var req protocol.DeletePaths
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(protocol.DELETE_FILES, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.DELETE_FILES_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.DeletePathsResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal DeleteFilesResp fail. err:", err)
		return
	}

	loggermsg.Info("delete files resp, resp:", resp)
}

func doRenamePath(rw *saferw.SafeRW, path, newName string) {
	var req protocol.RenamePath
	req.Nonce = rand.Uint32()
	req.Path = path
	req.Name = newName

	sendMsg(protocol.RENAME_PATH, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.RENAME_PATH_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.RenamePathResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal RenamePathResp fail. err:", err)
		return
	}

	loggermsg.Info("rename path resp, resp:", resp)
}

func doPutinRecycle(rw *saferw.SafeRW, paths []string) {
	var req protocol.PutInRecycle
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(protocol.PUTIN_RECYCLE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.PUTIN_RECYCLE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.PutInRecycleResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal PutInRecycleResp fail. err:", err)
		return
	}

	loggermsg.Info("putin recycle resp, resp:", resp)
}

func doDredgeOutRecycle(rw *saferw.SafeRW, paths []string) {
	var req protocol.DredgeOutRecycle
	req.Nonce = rand.Uint32()
	req.Paths = paths

	sendMsg(protocol.DREDGEOUT_RECYCLE, &req, rw)

	body := make([]byte, 1024*256)
	msgLen, msgId, body, err := protocol.ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read one msg fail. err:", err)
		return
	}

	if msgId != uint32(protocol.DREDGEOUT_RECYCLE_RESP) {
		loggermsg.Error("resp msg id invalid, msgid:", msgId)
		return
	}

	var resp protocol.DredgeOutRecycleResp
	err = proto.Unmarshal(body[:msgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal DredgeOutRecycleResp fail. err:", err)
		return
	}

	loggermsg.Info("dredge out recycle resp, resp:", resp)
}
