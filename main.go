package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"gitlab.dabank.io/nas/go-nas/common"
	"gitlab.dabank.io/nas/go-nas/p2p/protocol"
	"gitlab.dabank.io/nas/go-nas/saferw"
	"gitlab.dabank.io/nas/p2p-network/communication"
	"gitlab.dabank.io/nas/p2p-network/config"
	"gitlab.dabank.io/nas/p2p-network/core"
	"gitlab.dabank.io/nas/p2p-network/hosting"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"
)

type testTarget struct {
	hostPort int
	id       string
	ip       string
	port     int
	mode     config.PeerMode
	mapped   string
}

var (
	logger    = logging.Logger("p2p-network")
	loggermsg = logging.Logger("-----msg-----")

	udpClient *bool
	udpServer *bool

	taskid string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	logging.SetAllLoggers(logging.LevelInfo)
	//_ = logging.SetLogLevel("p2p-network", "DEBUG")
	//_ = logging.SetLogLevel("swarm2", "DEBUG")
	_ = logging.SetLogLevel("-----msg-----", "DEBUG")

	//bind()
	startP2p()
}

func startP2p() {
	args := readCmdArgs()

	if udpClient != nil && *udpClient {
		startUDPClient()
		return
	} else if udpServer != nil && *udpServer {
		startUDPServer()
		return
	}

	mockRouting := !core.IsBlankString(args.id) && !core.IsBlankString(args.ip) && args.port > 0

	dir, err := os.Getwd()
	panicIfError(err)

	logger.Info("Work dir: ", dir)

	fullPath := path.Join(dir, "private.key")

	if !core.FileExist(fullPath) {

		priKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
		panicIfError(err)
		content, err := priKey.Raw()
		panicIfError(err)

		err = ioutil.WriteFile(fullPath, content, os.FileMode(0722))
		panicIfError(err)
	}

	cnfBuilder := config.Builder().Logger(logger).Port(args.hostPort).PrivateKey(fullPath)
	cnfBuilder.Advertise(time.Second*10, time.Second*20)

	bootstrapPeers := []communication.PeerInfo{
		{
			"12D3KooWHGwQULtBt3VzbKFYvk5zHS24gZ4CYWgYzRgUJeVHkpo8",
			[]communication.EndPoint{
				{communication.AddressFormatLibp2p, "/dns4/dabank.coinbi.io/tcp/25556"},
			},
		},
	}
	cnfBuilder.KadRouting(bootstrapPeers...)

	var cnf config.HostConfig
	cnf = cnfBuilder.BuildPeer()

	host, err := hosting.NewPeer(context.Background(), cnf)
	if err != nil {
		panic(err)
	}

	<-host.Start()
	state, _ := host.GetState()
	logger.Info("----------host, peerid:", state.Id, ", address:", state.Addresses)
	loopDial(host, args)

	if mockRouting {
		logger.Info("Start Done")
	}

	select {}
}

func debugPeers(host hosting.NetworkPeer) string {
	sb := core.NewStringBuilder()
	sb.WriteLine("Known peers:")
	for _, p := range host.GetKnownPeers() {
		sb.WriteSprintLine("ID: ", p.Id)
		for _, endpoint := range p.Endpoints {
			sb.WriteLine(endpoint.Address)
		}
		sb.WriteLine("------------------------")
	}
	return sb.String()
}

func loopDial(host hosting.NetworkPeer, args *testTarget) {
	go func() {
		for {
			str := debugPeers(host)
			logger.Info(str)
			stream, err := host.Dial(args.id)
			logger.Info("~~~~~~~~~~~~~~~dial return")
			if err != nil {
				logger.Info("connect peer fault.", core.SysBreakLine(), "error: ", err.Error())

				time.Sleep(time.Second * 3)
				continue
			}
			//writeTestMessageLoop(stream, args)

			rw := saferw.CreateSafeRW(bufio.NewReader(stream), bufio.NewWriter(stream))
			onConnect(rw)
			return
		}
	}()
}

func readCmdArgs() *testTarget {
	///ip4/169.254.73.53/tcp/9718/p2p/12D3KooWNJQpY1e98JAqEB36FmPsVEtMFXa8PYshpuYmhS6HAZg1
	port := flag.Int("pp", 0, "target port")
	ip := flag.String("pip", "", "target address")
	id := flag.String("pid", "", "target id")

	hostPort := flag.Int("p", 0, "target id")

	pType := flag.String("m", "client", "peer type")

	mappedAddress := flag.String("l", "", "mapped ma address")

	udpClient = flag.Bool("udpc", false, "udp client")
	udpServer = flag.Bool("udps", false, "udp server")

	flag.Parse()

	var mode = config.PeerModePeer
	if *pType == "server" {
		mode = config.PeerModeBootstrap
	}

	return &testTarget{
		*hostPort, *id, *ip, *port, mode, *mappedAddress,
	}
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func doLogin(rw *saferw.SafeRW) error {
	var login protocol.AppLogin
	login.Nonce = rand.Uint32()
	login.Timestamp = uint64(time.Now().Unix())
	login.Token = "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff11"
	sendMsg(protocol.APP_LOGIN, &login, rw)

	loggermsg.Info("prepare read data")
	readMsg(rw)
	return nil
}

func readMsg(rw *saferw.SafeRW) {
	head := make([]byte, 8)
	rbuf := make([]byte, 1024*1024)
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

	dumpMsg(protocol.P2pMsgID(respCmd), rbuf[:respMsgLen-8])

	var body interface{}
	switch protocol.P2pMsgID(respCmd) {
	case protocol.PREPARE_UPLOADFILE_RESP:
		var preUploadResp protocol.PrepareUploadFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &preUploadResp)
		if err != nil {
			loggermsg.Error("protobuf unmarshal PrepareUploadFileResp fail. err:", err)
			return
		}

		taskid = preUploadResp.TaskId
		loggermsg.Info("upload task id:", taskid)
	case protocol.UPLOADFILE_RESULT:
		var uploadResult protocol.UploadFileResult
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &uploadResult)
		if err != nil {
			loggermsg.Error("protobuf unmarshal UploadFileResult fail. err:", err)
			return
		}

		loggermsg.Info("upload result:", uploadResult)

	case protocol.BOX_LOGIN_RESP:
		var loginResp protocol.BoxLoginResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &loginResp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal BoxLoginResp fail. err:", err)
			return
		}

		loggermsg.Info("box login result:", loginResp)

	case protocol.PREPARE_BACKUPFILE_RESP:
		var resp protocol.PrepareBackupFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal PrepareBackupFileResp fail. err:", err)
			return
		}

		taskid = resp.TaskId
		loggermsg.Info("prepare backup result:", resp)
	case protocol.PREPARE_RECOVERFILE_RESP:
		var resp protocol.PrepareRecoverFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal PrepareRecoverFileResp fail. err:", err)
			return
		}

		body = &resp
	}
	loggermsg.Info("received one msg, ", "cmd:", respCmd, ", body:", body)
}

func sendMsg(cmd protocol.P2pMsgID, msg proto.Message, rw *saferw.SafeRW) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		loggermsg.Error("protobuf marshal fail", ", cmd:", cmd, ", err:", err)
		return err
	}

	var head [8]byte
	SerialHead(uint32(len(body)+8), cmd, head[:])
	logger.Debug("send msg head", ", len:", len(body)+8, ", cmd:", cmd, ", head:", hex.EncodeToString(head[:]))
	n, err := rw.Write(head[:])
	if err != nil {
		loggermsg.Error("send msg fail", ", cmd:", cmd, ", err:", err)
		return err
	}

	if n != len(head) {
		loggermsg.Error("sent len doesn't match head len", ", n:", n)
		return common.ErrSendMsgHeadFail
	}

	loggermsg.Debug("sent head", ", n:", n)
	sn := 0
	for sn < len(body) {
		n, err := rw.Write(body[sn:])
		if err != nil {
			loggermsg.Error("send msg body fail", ", err:", err)
			return err
		}

		sn += n
		loggermsg.Debug("send loop", ", sn:", sn, ", n:", n)
	}

	err = rw.Flush()
	if err != nil {
		loggermsg.Error("connection flush fail", ", err:", err)
		return err
	}

	loggermsg.Debug("send msg successful", ", cmd:", cmd)
	return nil
}

func SerialHead(len uint32, cmd protocol.P2pMsgID, buf []byte) {
	binary.BigEndian.PutUint32(buf, len)
	binary.BigEndian.PutUint32(buf[4:], uint32(cmd))
}

func getFileHash(path string) []byte {
	hasher := sha256.New()
	f, _ := os.Open(path)
	defer f.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}

		hasher.Write(buf[:n])
	}

	return hasher.Sum(nil)
}

func getDataHash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func getFileSize(path string) uint64 {
	fi, _ := os.Stat(path)
	return uint64(fi.Size())
}

func dumpMsg(cmd protocol.P2pMsgID, body []byte) {
	switch cmd {
	case protocol.APP_LOGIN_RESP:
		var msg protocol.AppLoginResp
		proto.Unmarshal(body, &msg)
		fmt.Println("dump msg", "cmd", cmd, "body", msg)
	}
}

func serialHead(len uint32, cmd protocol.P2pMsgID, buf []byte) {
	binary.BigEndian.PutUint32(buf, len)
	binary.BigEndian.PutUint32(buf[4:], uint32(cmd))
	logger.Info("#################head:", hex.EncodeToString(buf))
}

func parseMsgLenAndCmd(buf []byte) (len uint32, cmd uint32) {
	len = binary.BigEndian.Uint32(buf)
	cmd = binary.BigEndian.Uint32(buf[4:])
	return
}

func getLoginReq() []byte {
	var msg protocol.AppLogin
	msg.Nonce = rand.Uint32()
	msg.Timestamp = uint64(time.Now().Unix())
	msg.Token = "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff11"

	data, _ := proto.Marshal(&msg)
	return data
}

func doBoxLogin(rw *saferw.SafeRW) {
	var login protocol.BoxLogin
	login.Nonce = rand.Uint32()
	login.Timestamp = uint64(time.Now().Unix())
	login.Token = ""

	time.Sleep(time.Second)
	err := sendMsg(protocol.BOX_LOGIN, &login, rw)
	if err != nil {
		loggermsg.Error("send box login msg fail, err:", err)
		return
	}

	readMsg(rw)
}

func onConnect(rw *saferw.SafeRW) {
	loggermsg.Info("############onConnect")
	//time.Sleep(time.Minute * 100)

	//{
	//	doLogin(rw)
	//	doCreateDir(rw, "/20201017-2/files/testdir1")
	//	doCreateDir(rw, "/20201017-2/files/testdir2")
	//	doCreateDir(rw, "/20201017-2/files/testdir3")
	//	doCreateDir(rw, "/20201017-2/files/testdir4")
	//	doCreateDir(rw, "/20201017-2/files/testdir5")
	//}

	//{
	//	doLogin(rw)
	//	//doExplorDir(rw, "/")
	//	//doExplorDir(rw, "/20201017-2")
	//	//doExplorDir(rw, "/20201017-2/")
	//	doExplorDir(rw, "/20201017-2/files/")
	//}

	//{
	//	doLogin(rw)
	//	doRenamePath(rw, "/20201017-2", "20201017-3")
	//}

	//{
	//	doLogin(rw)
	//
	//	paths := make([]string, 2)
	//	paths[0] = "/20201017-2/files/testdir4"
	//	paths[1] = "/20201017-2/files/testdir5"
	//	//doPutinRecycle(rw, paths)
	//	doDredgeOutRecycle(rw, paths)
	//}

	//{
	//	doLogin(rw)
	//	doRenamePath(rw, "/20201017-2", "20201017-3")
	//}

	//{
	//	doLogin(rw)
	//
	//	folderPaths := make([]string, 2)
	//	folderPaths[0] = "/20201017-3/files/testdir2"
	//	folderPaths[1] = "/20201017-3/files/testdir3"
	//	doDeleteFiles(rw, folderPaths)
	//
	//	//filePaths := make([]string, 2)
	//	//filePaths[0] = "/20201017/f/app1"
	//	//filePaths[1] = "/20201017/f/lotus_v0.1.0_linux-amd64.tar.gz"
	//	//doDeleteFiles(rw, filePaths)
	//}

	{
		remoteDir := "/20201019/files/20201020-4"
		doLogin(rw)
		loggermsg.Info("~~~~~~~~~~~~~~~~~~~~~~~~login suc")
		doUploadFile(rw, "lotus_v0.1.0_linux-amd64.tar.gz", remoteDir)
		//doUploadFile(rw, "lws-iot-sdk-master.zip", remoteDir)
		//doUploadFile(rw, "binding.go", remoteDir)
		//doUploadFile(rw, "app1", remoteDir)
		//doUploadFile(rw, "go.mod", remoteDir)
		//doUploadFile(rw, "go.sum", remoteDir)
		//doUploadFile(rw, "main.go", remoteDir)
		//doUploadFile(rw, "Makefile", remoteDir)
		//doUploadFile(rw, "private.key", remoteDir)
	}

	//{
	//	remoteDir := "/20201019/files/20201020-3/"
	//	doLogin(rw)
	//	time.Sleep(time.Minute * 10)
	//	doDownloadFile(rw, remoteDir+"lotus_v0.1.0_linux-amd64.tar.gz", "./20201020-1.dld")
	//	//doDownloadFile(rw, remoteDir+"lws-iot-sdk-master.zip", "./20201019-2.dld")
	//	//doDownloadFile(rw, remoteDir+"binding.go", "./20201019-3.dld")
	//	//doDownloadFile(rw, remoteDir+"app1", "./20201019-4.dld")
	//	//doDownloadFile(rw, remoteDir+"go.mod", "./20201019-5.dld")
	//	//doDownloadFile(rw, remoteDir+"go.sum", "./20201019-6.dld")
	//	//doDownloadFile(rw, remoteDir+"main.go", "./20201019-7.dld")
	//	//doDownloadFile(rw, remoteDir+"Makefile", "./20201019-8.dld")
	//	//doDownloadFile(rw, remoteDir+"private.key", "./20201019-9.dld")
	//}

	//{
	//	doBoxLogin(rw)
	//	doBackUpFile(rw, "./lotus_v0.1.0_linux-amd64.tar.gz")
	//}

	//{
	//	doBoxLogin(rw)
	//	doRecover(rw, "b245e99e88745c6efbf7af7cf2f0f8e65faa05fe59a2a678485256bcabb86d42")
	//}

	//{
	//	doLogin(rw)
	//	doReset(rw)
	//}

}
