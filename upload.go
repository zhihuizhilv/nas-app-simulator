package main

import (
	"bytes"
	"crypto/sha256"
	"gitlab.dabank.io/nas/go-nas/p2p/protocol"
	"gitlab.dabank.io/nas/go-nas/saferw"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"time"
)

func doUploadFile(rw *saferw.SafeRW, filename string, dir string) {
	filePath := "./" + filename
	fileSize := getFileSize(filePath)

	var prepareUpload protocol.PrepareUploadFile
	prepareUpload.Nonce = rand.Uint32()
	prepareUpload.Path = dir + "/" + filename
	prepareUpload.Size = fileSize
	prepareUpload.Hash = getFileHash(filePath)
	prepareUpload.Encrypt = 0
	prepareUpload.BackupNum = 1

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE")
	err := sendMsg(protocol.PREPARE_UPLOADFILE, &prepareUpload, rw)
	if err != nil {
		return
	}

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE end")
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
	buf := make([]byte, frameSize)
	for {
		n, _ := f.Read(buf)
		if n == 0 {
			break
		}

		var uploadFrame protocol.UploadFileFrame
		uploadFrame.TaskId = taskid
		uploadFrame.FrameNum = uint32(frameNum)
		uploadFrame.FrameId = i
		uploadFrame.FrameSize = uint32(n)
		uploadFrame.FrameHash = getDataHash(buf[:n])
		uploadFrame.Data = buf[:n]
		sendMsg(protocol.UPLOADFILE_FRAME, &uploadFrame, rw)
		loggermsg.Info("sent frame. frame id:", i, ", len:", n)

		i++
	}

	loggermsg.Info("sent frames finish")
	readMsg(rw)
}

func doDownloadFile(rw *saferw.SafeRW, path, name string) {
	srcPath := path
	desPath := name

	var preDownload protocol.PrepareDownloadFile
	preDownload.Nonce = rand.Uint32()
	preDownload.Path = srcPath
	preDownload.Offset = 0

	err := sendMsg(protocol.PREPARE_DOWNLOADFILE, &preDownload, rw)
	if err != nil {
		return
	}

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
	if protocol.P2pMsgID(respCmd) != protocol.PREPARE_DOWNLOADFILE_RESP {
		loggermsg.Error("invalid prepare download resp msg cmd")
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

	var preDownloadResp protocol.PrepareDownloadFileResp
	err = proto.Unmarshal(rbuf[:respMsgLen-8], &preDownloadResp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal PrepareDownloadFileResp fail. err:", err)
		return
	}

	loggermsg.Info("get prepare download resp:", preDownloadResp)
	if preDownloadResp.RetCode != 0 {
		loggermsg.Error("prepare download file fail. err:", preDownloadResp.Description)
		return
	}

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
		if protocol.P2pMsgID(respCmd) != protocol.DOWNLOAD_FILEFRAME {
			loggermsg.Error("invalid download frame msg cmd")
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

		var downloadFramd protocol.DownloadFileFrame
		err = proto.Unmarshal(rbuf[:respMsgLen-8], &downloadFramd)
		if err != nil {
			loggermsg.Error("protobuf unmarshal DownloadFileFrame fail. err:", err)
			return
		}

		hasher := sha256.New()
		hasher.Write(downloadFramd.Data)
		hash := hasher.Sum(nil)
		if bytes.Compare(hash, downloadFramd.FrameHash) != 0 {
			loggermsg.Error("frame data hash dont match!!!!!!!!!!!! frameId:", downloadFramd.FrameId)
		}

		loggermsg.Info("received a download frame. frameNum:", downloadFramd.FrameNum, ", frameId:", downloadFramd.FrameId, ", dataLen:", len(downloadFramd.Data), ", path:", downloadFramd.Path)
		wn, err := f.Write(downloadFramd.Data)
		if err != nil {
			loggermsg.Error("write data into file fail. err:", err)
			return
		}

		loggermsg.Info("write data into file. wn:", wn)
		if downloadFramd.FrameId == downloadFramd.FrameNum-1 {
			loggermsg.Info("download file complete")
			return
		}

		////////////////////////////////////test/////////////////////////////////////
		//time.Sleep(time.Second)
		////////////////////////////////////test/////////////////////////////////////
	}
}
