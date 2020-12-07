package main

import (
	"bytes"
	"crypto/sha256"
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/saferw"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"math/rand"
	"os"
)

func doUploadFile(rw *saferw.SafeRW, filename string, dir string) {
	loggermsg.Info("upload faile begin ~~~~~~~~~")
	defer loggermsg.Info("upload faile end ~~~~~~~~~")

	filePath := "./" + filename
	fileSize := getFileSize(filePath)

	var prepareUpload p2pprotocol.PrepareUploadFile
	prepareUpload.Nonce = rand.Uint32()
	prepareUpload.Path = dir + "/" + filename
	prepareUpload.MimeType = "application/octet-stream"
	prepareUpload.Size = fileSize
	prepareUpload.Hash = getFileHash(filePath)
	prepareUpload.Encrypt = 0
	prepareUpload.BackupNum = 1
	prepareUpload.BackupTerm = 100000


	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE:", prepareUpload)
	err := sendMsg(p2pprotocol.PREPARE_UPLOADFILE, &prepareUpload, rw)
	if err != nil {
		return
	}

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE end")
	body := make([]byte, 1024)
	len, cmd, body, err := ReadOneMsg(rw, body)
	if p2pprotocol.P2pMsgID(cmd) != p2pprotocol.PREPARE_UPLOADFILE_RESP {
		loggermsg.Error("invalid resp msg id")
		return
	}

	var resp p2pprotocol.PrepareUploadFileResp
	err = proto.Unmarshal(body[:len-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal fail. err:", err)
		return
	}

	loggermsg.Info("prepare upload resp:", resp)
	loggermsg.Info("~~~~~~~~~~upload offset:", resp.Offset)

	// 发送数据帧
	f, err := os.Open(filePath)
	if err != nil {
		loggermsg.Error("open file fail, err:", err)
		return
	}
	defer f.Close()

	f.Seek(int64(resp.Offset), 0)
	frameSize := uint64(1024 * 1024)
	frameNum := (fileSize - resp.Offset) / frameSize
	if fileSize%frameSize != 0 {
		frameNum++
	}

	loggermsg.Info("fileSize:", fileSize, ", frameSize:", frameSize, ", frameNum:", frameNum)
	var i uint32
	offset := resp.Offset
	buf := make([]byte, frameSize)
	for {
		n, _ := f.Read(buf)
		if n == 0 {
			break
		}

		var uploadFrame p2pprotocol.UploadFileFrame
		uploadFrame.TaskId = resp.TaskId
		uploadFrame.FrameNum = uint32(frameNum)
		uploadFrame.FrameId = i
		uploadFrame.Offset = offset
		uploadFrame.FrameHash = getDataHash(buf[:n])
		uploadFrame.Data = buf[:n]
		sendMsg(p2pprotocol.UPLOADFILE_FRAME, &uploadFrame, rw)
		loggermsg.Info("sent frame. frame id:", i, ", len:", n)

		i++
		offset += uint64(n)
	}

	loggermsg.Info("sent frames finish")
	readMsg(rw)
}


func doUploadFileHalf(rw *saferw.SafeRW, filename string, dir string) {
	loggermsg.Info("upload faile begin ~~~~~~~~~")
	defer loggermsg.Info("upload faile end ~~~~~~~~~")

	filePath := "./" + filename
	fileSize := getFileSize(filePath)

	var prepareUpload p2pprotocol.PrepareUploadFile
	prepareUpload.Nonce = rand.Uint32()
	prepareUpload.Path = dir + "/" + filename
	prepareUpload.MimeType = "application/octet-stream"
	prepareUpload.Size = fileSize
	prepareUpload.Hash = getFileHash(filePath)
	prepareUpload.Encrypt = 0
	prepareUpload.BackupNum = 0

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE")
	err := sendMsg(p2pprotocol.PREPARE_UPLOADFILE, &prepareUpload, rw)
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
	var offset uint64
	buf := make([]byte, frameSize)
	for {
		n, _ := f.Read(buf)
		if n == 0 {
			break
		}

		var uploadFrame p2pprotocol.UploadFileFrame
		uploadFrame.TaskId = taskid
		uploadFrame.FrameNum = uint32(frameNum)
		uploadFrame.FrameId = i
		uploadFrame.Offset = offset
		uploadFrame.FrameHash = getDataHash(buf[:n])
		uploadFrame.Data = buf[:n]
		sendMsg(p2pprotocol.UPLOADFILE_FRAME, &uploadFrame, rw)
		loggermsg.Info("sent frame. frame id:", i, ", len:", n)

		i++

		if i >= uint32(40) {
			return
		}
		offset += uint64(n)
	}

	loggermsg.Info("sent frames finish")
	readMsg(rw)
}

func doUploadThumbnail(rw *saferw.SafeRW, filename, thumbnail, dir string) {
	loggermsg.Info("upload faile begin ~~~~~~~~~")
	defer loggermsg.Info("upload faile end ~~~~~~~~~")

	filePath := "./" + filename
	fileSize := getFileSize(filePath)

	///////////////////////////////prepare//////////////////////////////////
	var prepareUpload p2pprotocol.PrepareUploadFile
	prepareUpload.Nonce = rand.Uint32()
	prepareUpload.Path = dir + "/" + filename
	prepareUpload.Size = fileSize
	prepareUpload.Hash = getFileHash(filePath)
	prepareUpload.Encrypt = 1
	prepareUpload.BackupNum = 0

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE")
	err := sendMsg(p2pprotocol.PREPARE_UPLOADFILE, &prepareUpload, rw)
	if err != nil {
		return
	}

	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE end")
	readMsg(rw)
	loggermsg.Info("begin send file frame")
	///////////////////////////////prepare//////////////////////////////////


	///////////////////////////////thumbnail//////////////////////////////////
	var uploadThumbnail p2pprotocol.UploadFileThumbnail
	data, err := ioutil.ReadFile(thumbnail)
	if err != nil {
		loggermsg.Error("read thumbnail fail. path:", thumbnail, ", err:", err)
		return
	}

	uploadThumbnail.TaskId = taskid
	uploadThumbnail.Data = data
	loggermsg.Info("~~~~~~~~~send msg : PREPARE_UPLOADFILE")
	err = sendMsg(p2pprotocol.UPLOADFILE_THUMBNAIL, &uploadThumbnail, rw)
	if err != nil {
		return
	}
	///////////////////////////////thumbnail//////////////////////////////////


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

		var uploadFrame p2pprotocol.UploadFileFrame
		uploadFrame.TaskId = taskid
		uploadFrame.FrameNum = uint32(frameNum)
		uploadFrame.FrameId = i
		uploadFrame.Offset = offset
		uploadFrame.FrameHash = getDataHash(buf[:n])
		uploadFrame.Data = buf[:n]
		sendMsg(p2pprotocol.UPLOADFILE_FRAME, &uploadFrame, rw)
		loggermsg.Info("sent frame. frame id:", i, ", len:", n)

		i++
		offset += uint64(n)
	}

	loggermsg.Info("sent frames finish")
	readMsg(rw)
}

func doDownloadThumbnail(rw *saferw.SafeRW, path, name string)  {
	srcPath := path
	desPath := name

	var req p2pprotocol.DownloadThumbnail
	req.Nonce = rand.Uint32()
	req.Path = srcPath

	err := sendMsg(p2pprotocol.DOWNLOAD_THUMBNAIL, &req, rw)
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
	if p2pprotocol.P2pMsgID(respCmd) != p2pprotocol.DOWNLOAD_THUMBNAIL_RESP {
		loggermsg.Error("invalid download thumbnail resp msg cmd")
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

	var resp p2pprotocol.DownloadThumbnailResp
	err = proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal DownloadThumbnailResp fail. err:", err)
		return
	}

	loggermsg.Info("get download thumbnial resp:", resp)
	if resp.RetCode != 0 {
		loggermsg.Error("download thumbnail file fail. err:", resp.Description)
		return
	}

	f, err := os.Create(desPath)
	if err != nil {
		loggermsg.Error("create file fail, err:", err)
		return
	}
	defer f.Close()
	f.Write(resp.Data)
}

func doDownloadFile(rw *saferw.SafeRW, path, name string) {
	srcPath := path
	desPath := name

	var preDownload p2pprotocol.PrepareDownloadFile
	preDownload.Nonce = rand.Uint32()
	preDownload.Path = srcPath
	preDownload.Offset = getFileSize(desPath) - 1024

	err := sendMsg(p2pprotocol.PREPARE_DOWNLOADFILE, &preDownload, rw)
	if err != nil {
		return
	}

	body := make([]byte, 1024*1024*2)
	msglen, cmd, body, err := ReadOneMsg(rw, body)
	if p2pprotocol.P2pMsgID(cmd) != p2pprotocol.PREPARE_DOWNLOADFILE_RESP {
		loggermsg.Error("invalid prepare download resp msg cmd")
		return
	}

	var resp p2pprotocol.PrepareDownloadFileResp
	err = proto.Unmarshal(body[:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("protobuf unmarshal PrepareDownloadFileResp fail. err:", err)
		return
	}

	loggermsg.Info("get prepare download resp:", resp)
	if resp.RetCode != 0 {
		loggermsg.Error("prepare download file fail. err:", resp.Description)
		return
	}

	loggermsg.Info("~~~offset:", resp.Offset)


	f, err := os.OpenFile(desPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		loggermsg.Error("create file fail, err:", err)
		return
	}
	defer f.Close()
	f.Seek(int64(resp.Offset), 0)

	for {
		msglen, cmd, body, err = ReadOneMsg(rw, body)
		loggermsg.Info("msg len:", msglen, ", msg cmd:", cmd)
		if p2pprotocol.P2pMsgID(cmd) != p2pprotocol.DOWNLOAD_FILEFRAME {
			loggermsg.Error("invalid download frame msg cmd")
			return
		}

		var downloadFramd p2pprotocol.DownloadFileFrame
		err = proto.Unmarshal(body[:msglen-8], &downloadFramd)
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

		//if downloadFramd.FrameId == 34 {
		//	return
		//}
	}
}
