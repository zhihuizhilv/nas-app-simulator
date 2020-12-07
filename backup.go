package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"time"
)

func doBackUpFile(rw *saferw.SafeRW, name string) {
	filePath := name
	fileSize := getFileSize(filePath)

	bbcAddr := "15wzfexznydjrsr0qwfma2ppjqs3y1krkpbdvhkme81y7wemytcaf8k28"

	var prepare p2pprotocol.PrepareBackupFile
	prepare.Nonce = rand.Uint32()
	prepare.Timestamp = uint64(time.Now().Unix())
	prepare.BbcAddr = bbcAddr
	prepare.Size = fileSize
	prepare.Hash = getFileHash(filePath)
	prepare.Price = 10000
	prepare.Signature = nil

	err := sendMsg(p2pprotocol.PREPARE_BACKUPFILE, &prepare, rw)
	if err != nil {
		return
	}

	readMsg(rw)
	loggermsg.Info("begin send file frame")

	// 发送数据帧
	f, err := os.Open(filePath)
	if err != nil {
		loggermsg.Error("open file fail, err:", err)
		return
	}
	defer f.Close()

	frameSize := uint64(1024 * 1024)
	frameNum := fileSize / frameSize
	if fileSize%frameSize != 0 {
		frameNum++
	}

	loggermsg.Info("fileSize:", fileSize, ", frameSize:", frameSize, ", frameNum:", frameNum)
	var i uint32
	var offset uint64
	buf := make([]byte, frameSize)
	for {
		n, _ := f.Read(buf)
		if n == 0 {
			break
		}

		var frame p2pprotocol.BackupFileFrame
		frame.TaskId = taskid
		frame.FrameNum = uint32(frameNum)
		frame.FrameId = i
		frame.Offset = offset
		frame.FrameHash = getDataHash(buf[:n])
		frame.Data = buf[:n]
		sendMsg(p2pprotocol.BACKUPFILE_FRAME, &frame, rw)
		loggermsg.Info("sent frame. frame id:", i, ", len:", n)

		i++
		offset += uint64(n)
	}

	loggermsg.Info("sent frames finish")
	readMsg(rw)
}

func doRecover(rw *saferw.SafeRW, hashStr string) {
	bbcAddr := "15wzfexznydjrsr0qwfma2ppjqs3y1krkpbdvhkme81y7wemytcaf8k28"
	hash, _ := hex.DecodeString(hashStr)
	sign, _ := hex.DecodeString("112233")

	var prepare p2pprotocol.PrepareRecoverFile
	prepare.Nonce = rand.Uint32()
	prepare.Timestamp = uint64(time.Now().Unix())
	prepare.BbcAddr = bbcAddr
	prepare.Hash = hash
	prepare.Offset = 0
	prepare.Signature = sign

	sendMsg(p2pprotocol.PREPARE_RECOVERFILE, &prepare, rw)

	rbufLen := 10 * 1024 * 1024
	head := make([]byte, 8)
	rbuf := make([]byte, rbufLen)
	rn, err := rw.Read(head)
	if err != nil {
		loggermsg.Error("read msg head fail. err:", err)
		return
	}
	loggermsg.Info("read head len:", rn)

	if rn != len(head) {
		loggermsg.Error("read msg head fail. n:", rn)
		return
	}

	respMsgLen, respCmd := parseMsgLenAndCmd(head)
	loggermsg.Info("msg len:", respMsgLen, ", msg cmd:", respCmd)
	if p2pprotocol.P2pMsgID(respCmd) != p2pprotocol.PREPARE_RECOVERFILE_RESP {
		loggermsg.Error("invalid prepare recover resp msg cmd")
		return
	}

	rn = 0
	for rn < int(respMsgLen-8) {
		n, err := rw.Read(rbuf[rn:])
		if err != nil {
			loggermsg.Error("receive msg fail, err:", err)
			return
		}

		rn += n
		loggermsg.Info("read msg loop, n:", n, ", rn:", rn)
	}

	var prepareResp p2pprotocol.PrepareRecoverFileResp
	err = proto.Unmarshal(rbuf[:respMsgLen-8], &prepareResp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal PrepareRecoverFileResp fail. err:", err)
		return
	}

	loggermsg.Info("get prepare recover resp:", prepareResp)
	if prepareResp.RetCode != 0 {
		loggermsg.Error("prepare recover file fail. err:", prepareResp.Description)
		return
	}

	desPath := "./recover.dld"
	f, err := os.Create(desPath)
	if err != nil {
		loggermsg.Error("create file fail, err:", err)
		return
	}
	defer f.Close()

	for {
		rn, err := rw.Read(head)
		if err != nil {
			loggermsg.Error("read msg head fail. err:", err)
			return
		}
		loggermsg.Info("read head len:", rn)

		if rn != len(head) {
			loggermsg.Error("read msg head fail. n:", rn)
			return
		}

		respMsgLen, respCmd := parseMsgLenAndCmd(head)
		loggermsg.Info("msg len:", respMsgLen, ", msg cmd:", respCmd)
		if p2pprotocol.P2pMsgID(respCmd) != p2pprotocol.RECOVERFILE_FRAME {
			loggermsg.Error("invalid recover frame msg cmd")
			return
		}

		rn = 0
		for rn < int(respMsgLen-8) {
			n, err := rw.Read(rbuf[rn : respMsgLen-8])
			if err != nil {
				loggermsg.Error("receive msg fail, err:", err)
				return
			}

			rn += n
			loggermsg.Info("read msg loop, n:", n, ", rn:", rn)
			if n == 0 {
				time.Sleep(time.Second)
			}
		}

		var frame p2pprotocol.RecoverFileFrame
		err = proto.Unmarshal(rbuf[:respMsgLen-8], &frame)
		if err != nil {
			loggermsg.Error("protobuf unmarshal RecoverFileFrame fail. err:", err)
			return
		}

		hasher := sha256.New()
		hasher.Write(frame.Data)
		hash := hasher.Sum(nil)
		if bytes.Compare(hash, frame.FrameHash) != 0 {
			loggermsg.Error("frame data hash dont match!!!!!!!!!!!! frameId:", frame.FrameId)
		}

		loggermsg.Info("received a recover frame. frameNum:", frame.FrameNum, ", frameId:", frame.FrameId, ", dataLen:", len(frame.Data))
		wn, err := f.Write(frame.Data)
		if err != nil {
			loggermsg.Error("write data into file fail. err:", err)
			return
		}

		loggermsg.Info("write data into file. wn:", wn)
		if frame.FrameId == frame.FrameNum-1 {
			loggermsg.Info("recover file complete")
			return
		}
	}
}

func doChallange(rw *saferw.SafeRW) {
	var req p2pprotocol.StorageChallenge
	req.Nonce = rand.Uint32()
	req.Timestamp = uint64(time.Now().Unix())
	req.BbcAddr = "1dj09mvxt72x70bnqnpt25xf69hgfz1d5n40tjn1gv1w6dts8ds8tm9bj"
	req.Name = "42c24694-9cab-41df-87e9-3bfb8a9cd368"
	req.Position = rand.Uint64()%31061700

	sendMsg(p2pprotocol.STORAGE_CHALLENGE, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.STORAGE_CHALLENGE_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.STORAGE_CHALLENGE_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.StorageChallengeResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal StorageChallengeResp fail, err:", err)
		return
	}

	loggermsg.Info("StorageChallengeResp:", resp)
}