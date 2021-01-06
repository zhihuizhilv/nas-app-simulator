package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"gitlab.dabank.io/nas/app-simulator/bbcsign"
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/saferw"
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
	MAXMSGLEN = 1024 * 1024 * 2
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

	//mockRouting := !core.IsBlankString(args.id) && !core.IsBlankString(args.ip) && args.port > 0

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

	//bootstrapPeers := []communication.PeerInfo{
	//	{
	//		"12D3KooWHGwQULtBt3VzbKFYvk5zHS24gZ4CYWgYzRgUJeVHkpo8",
	//		[]communication.EndPoint{
	//			{communication.AddressFormatLibp2p, "/dns4/dabank.coinbi.io/tcp/25556"},
	//		},
	//	},
	//}
	bootstrapPeers := []communication.PeerInfo{
		{
			"12D3KooWNbhT78jpWBUvKuyEVJJhkwSDkMgfa187YD6ijepkQWg3",
			[]communication.EndPoint{
				{communication.AddressFormatLibp2p, "/dns4/dabank.coinbi.io/tcp/25557"},
			},
		},
	}
	cnfBuilder.KadRouting(bootstrapPeers...)

	var cnf config.HostConfig
	cnf = cnfBuilder.BuildPeer()
	//cnf = cnfBuilder.BuildClient()
	fmt.Println("create client host config. cnf:", cnf)

	host, err := hosting.NewPeer(context.Background(), cnf)
	if err != nil {
		panic(err)
	}

	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~create peer suc")

	<-host.Start()
	state, _ := host.GetState()
	logger.Info("----------host, peerid:", state.Id, ", address:", state.Addresses)
	loopDial(host, args)

	//if mockRouting {
	//	logger.Info("Start Done")
	//}

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

func GetAppLoginSignData(login *p2pprotocol.AppLogin) []byte {
	data := fmt.Sprintf("%d%d%s", login.Nonce, login.Timestamp, login.BbcAddr)
	return []byte(data)
}


func doLogin(rw *saferw.SafeRW) error {
	var login p2pprotocol.AppLogin
	login.Nonce = rand.Uint32()
	login.Timestamp = uint64(time.Now().Unix())
	login.BbcAddr = BbcAddr

	signData := GetAppLoginSignData(&login)
	signBuf := bbcsign.Ed25519Sign(BbcKey, signData)
	signBuf[0] += 1
	login.Sign = hex.EncodeToString(signBuf)
	sendMsg(p2pprotocol.APP_LOGIN, &login, rw)

	loggermsg.Info("prepare read data")
	readMsg(rw)
	return nil
}

func ReadOneMsg(rw *saferw.SafeRW, body []byte) (uint32, uint32, []byte, error) {
	var msgHead [8]byte
	n, err := rw.Read(msgHead[:])
	if err != nil {
		return 0, 0, body, err
	}

	if n != len(msgHead) {
		return 0, 0, body, errors.New("read msg head fail")
	}

	msgLen, msgCmd := parseMsgLenAndCmd(msgHead[:])
	if msgLen > uint32(MAXMSGLEN) {
		return 0, 0, body, errors.New("invalid msg len")
	}

	if msgLen > uint32(len(body)) {
		body = make([]byte, msgLen)
	}

	bodyLen := 0
	for uint32(bodyLen) < msgLen-8 {
		n, err := rw.Read(body[bodyLen:])
		if err != nil {
			return 0, 0, body, err
		}

		bodyLen += n
	}

	return msgLen, msgCmd, body, nil
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

	dumpMsg(p2pprotocol.P2pMsgID(respCmd), rbuf[:respMsgLen-8])

	var body interface{}
	switch p2pprotocol.P2pMsgID(respCmd) {
	case p2pprotocol.PREPARE_UPLOADFILE_RESP:
		var preUploadResp p2pprotocol.PrepareUploadFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &preUploadResp)
		if err != nil {
			loggermsg.Error("protobuf unmarshal PrepareUploadFileResp fail. err:", err)
			return
		}

		taskid = preUploadResp.TaskId
		loggermsg.Info("prepare upload resp:", preUploadResp)
	case p2pprotocol.UPLOADFILE_RESULT:
		var uploadResult p2pprotocol.UploadFileResult
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &uploadResult)
		if err != nil {
			loggermsg.Error("protobuf unmarshal UploadFileResult fail. err:", err)
			return
		}

		loggermsg.Info("upload result:", uploadResult)

	case p2pprotocol.BOX_LOGIN_RESP:
		var loginResp p2pprotocol.BoxLoginResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &loginResp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal BoxLoginResp fail. err:", err)
			return
		}

		loggermsg.Info("box login result:", loginResp)

	case p2pprotocol.PREPARE_BACKUPFILE_RESP:
		var resp p2pprotocol.PrepareBackupFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal PrepareBackupFileResp fail. err:", err)
			return
		}

		taskid = resp.TaskId
		loggermsg.Info("prepare backup result:", resp)
	case p2pprotocol.PREPARE_RECOVERFILE_RESP:
		var resp p2pprotocol.PrepareRecoverFileResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal PrepareRecoverFileResp fail. err:", err)
			return
		}

		body = &resp
	case p2pprotocol.GET_BOXSTATUS_RESP:
		var resp p2pprotocol.GetStatusResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal GetStatusResp fail. err:", err)
			return
		}

		body = &resp
	case p2pprotocol.LIST_RECYCLE_RESP:

		var resp p2pprotocol.ListRecycleResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal ListRecycleResp fail. err:", err)
			return
		}

		body = &resp
	case p2pprotocol.SPACE_SETTING_RESP:

		var resp p2pprotocol.SpaceSettingResp
		err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
		if err != nil {
			loggermsg.Error("prorobuf unmarshal SpaceSettingResp fail. err:", err)
			return
		}

		body = &resp
	//case p2pprotocol.UPLOADFILE_THUMBNAIL:
	//
	//	var resp p2pprotocol.SpaceSettingResp
	//	err := proto.Unmarshal(rbuf[:respMsgLen-8], &resp)
	//	if err != nil {
	//		loggermsg.Error("prorobuf unmarshal SpaceSettingResp fail. err:", err)
	//		return
	//	}
	//
	//	body = &resp
	}

	loggermsg.Info("received one msg, ", "cmd:", respCmd, ", body:", body)
}

func sendMsg(cmd p2pprotocol.P2pMsgID, msg proto.Message, rw *saferw.SafeRW) error {
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
		return p2pprotocol.ErrSendMsgHeadFail
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

func SerialHead(len uint32, cmd p2pprotocol.P2pMsgID, buf []byte) {
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
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return uint64(fi.Size())
}

func dumpMsg(cmd p2pprotocol.P2pMsgID, body []byte) {
	switch cmd {
	case p2pprotocol.APP_LOGIN_RESP:
		var msg p2pprotocol.AppLoginResp
		proto.Unmarshal(body, &msg)
		fmt.Println("dump msg", "cmd", cmd, "body", msg)
	}
}

func serialHead(len uint32, cmd p2pprotocol.P2pMsgID, buf []byte) {
	binary.BigEndian.PutUint32(buf, len)
	binary.BigEndian.PutUint32(buf[4:], uint32(cmd))
	logger.Info("#################head:", hex.EncodeToString(buf))
}

func parseMsgLenAndCmd(buf []byte) (len uint32, cmd uint32) {
	len = binary.BigEndian.Uint32(buf)
	cmd = binary.BigEndian.Uint32(buf[4:])
	return
}

func doBoxLogin(rw *saferw.SafeRW) {
	var login p2pprotocol.BoxLogin
	login.Nonce = rand.Uint32()
	login.Timestamp = uint64(time.Now().Unix())
	login.Token = ""

	err := sendMsg(p2pprotocol.BOX_LOGIN, &login, rw)
	if err != nil {
		loggermsg.Error("send box login msg fail, err:", err)
		return
	}

	readMsg(rw)
}

func doGetState(rw *saferw.SafeRW) {
	var login p2pprotocol.GetStatus
	login.Nonce = rand.Uint32()

	err := sendMsg(p2pprotocol.GET_BOXSTATUS, &login, rw)
	if err != nil {
		loggermsg.Error("send get box status msg fail, err:", err)
		return
	}

	readMsg(rw)
}

func doListRecycle(rw *saferw.SafeRW) {
	var login p2pprotocol.ListRecycle
	login.Nonce = rand.Uint32()

	err := sendMsg(p2pprotocol.LIST_RECYCLE, &login, rw)
	if err != nil {
		loggermsg.Error("send list recycle msg fail, err:", err)
		return
	}

	readMsg(rw)
}

func doSpaceSetting(rw *saferw.SafeRW) {
	var setting p2pprotocol.SpaceSetting
	setting.Nonce = rand.Uint32()
	setting.ReservedSpace = 1024*1024*1024*10
	setting.SharedSpace = 296092692480 - setting.ReservedSpace
	err := sendMsg(p2pprotocol.SPACE_SETTING, &setting, rw)
	if err != nil {
		loggermsg.Error("send space setting fail. err:", err)
		return
	}

	readMsg(rw)
}

func onConnect(rw *saferw.SafeRW) {
	loggermsg.Info("############onConnect")
	loggermsg.Info("working start~~~~~~~~~~~~")
	//time.Sleep(time.Minute * 100)

	{
		doLogin(rw)
		doGetState(rw)
	}

	//{
	//	doLogin(rw)
	//	doGetState(rw)
	//	doSpaceSetting(rw)
	//	doGetState(rw)
	//}

	//{
	//	doLogin(rw)
	//	doCreateDir(rw, "/20201113/lzh1/testdir1")
	//	doCreateDir(rw, "/20201113/lzh1/testdir2")
	//	doCreateDir(rw, "/20201113/lzh1/testdir3")
	//	doCreateDir(rw, "/20201113/lzh1/testdir4")
	//	doCreateDir(rw, "/20201113/lzh1/testdir5")
	//}

	//{
	//	doLogin(rw)
	//	doAppendBackupTerm(rw, "/20201216/lzh1/app_dld1", 100000)
	//	//doExplorDir(rw, "/20201017-2")
	//	//doExplorDir(rw, "/20201017-2/")
	//	//doExplorDir(rw, "/20201103/lzh1/")
	//	//doExplorDir(rw, "/20201102/lzh1/data")
	//}

	//{
	//	//doLogin(rw)
	//	//doExplorDir(rw, "/20201123/lzh1")
	//	//doExplorDir(rw, "/20201017-2")
	//	//doExplorDir(rw, "/20201017-2/")
	//	//doExplorDir(rw, "/20201103/lzh1/")
	//	//doExplorDir(rw, "/20201102/lzh1/data")
	//}



	//{
	//	doLogin(rw)
	//	doRenamePath(rw, "/20201017-2", "20201017-3")
	//}

	//{
	//	doLogin(rw)
	//
	//	paths := make([]string, 2)
	//	paths[0] = "/20201103/lzh1/private.key"
	//	doPutinRecycle(rw, paths)
	//	//doDredgeOutRecycle(rw, paths)
	//}
	//
	//{
	//	doLogin(rw)
	//	doListRecycle(rw)
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

	//{
	//	remoteDir := "/20201223/lzh1"
	//	doLogin(rw)
	//	//doUploadFileHalf(rw, "data/lotus_v0.1.0_linux-amd64.tar.gz", remoteDir)
	//	//doUploadFile(rw, "data/lotus_v0.1.0_linux-amd64.tar.gz", remoteDir)
	//	//doUploadFile(rw, "data/lws-iot-sdk-master.zip", remoteDir)
	//	//doUploadFile(rw, "data/app1", remoteDir)
	//	//doUploadFile(rw, "go.mod", remoteDir)
	//	//doUploadFile(rw, "go.sum", remoteDir)
	//	//doUploadFile(rw, "main.go", remoteDir)
	//	doUploadFile(rw, "app_dld1", remoteDir)
	//	//doUploadFile(rw, "app_dld2", remoteDir)
	//}


	//{
	//	remoteDir := "/20201104/lzh2/"
	//	doLogin(rw)
	//	doUploadFileHalf(rw, "data/app1", remoteDir)
	//}


	//{
	//	remoteDir := "/20201028/lzh1/"
	//	doLogin(rw)
	//	doUploadThumbnail(rw, "data/app1", "data/thumbnail", remoteDir)
	//}

	//{
	//	remoteDir := "/20201028/lzh1/"
	//	doLogin(rw)
	//	doDownloadThumbnail(rw, remoteDir+"data/app1", "./dld2/20201028-1.dld")
	//}


	//{
	//	remoteDir := "/20201104/lzh3/"
	//	doLogin(rw)
	//	doDownloadFile(rw, remoteDir+"data/lotus_v0.1.0_linux-amd64.tar.gz", "./dld2/20201104-1.dld")
	//	//doDownloadFile(rw, remoteDir+"data/lws-iot-sdk-master.zip", "./dld2/20201023-2.dld")
	//	//doDownloadFile(rw, remoteDir+"data/app1", "./dld2/20201023-3.dld")
	//	//doDownloadFile(rw, remoteDir+"go.mod", "./dld2/20201023-4.dld")
	//	//doDownloadFile(rw, remoteDir+"go.sum", "./dld2/20201023-5.dld")
	//	//doDownloadFile(rw, remoteDir+"main.go", "./dld2/20201023-6.dld")
	//	//doDownloadFile(rw, remoteDir+"Makefile", "./dld2/20201023-7.dld")
	//	//doDownloadFile(rw, remoteDir+"private.key", "./dld2/20201023-8.dld")
	//}

	//{
	//	doBoxLogin(rw)
	//	doBackUpFile(rw, "./lotus_v0.1.0_linux-amd64.tar.gz")
	//}

	//{
	//	doBoxLogin(rw)
	//	doChallange(rw)
	//}

	//{
	//	doBoxLogin(rw)
	//	doRecover(rw, "b245e99e88745c6efbf7af7cf2f0f8e65faa05fe59a2a678485256bcabb86d42")
	//}

	//{
	//	doLogin(rw)
	//	doReset(rw)
	//}

	//{
	//	doLogin(rw)
	//	doGetLifeCycle(rw)
	//	for {
	//		doGetRecoverProgress(rw)
	//		time.Sleep(time.Second*2)
	//	}
	//}

	//{
	//	doLogin(rw)
	//	doGetFileInfo(rw, "/20201207")
	//	doGetFileInfo(rw, "/app-release.apkqqq")
	//}

	//{
	//	doLogin(rw)
	//	doOpenBinding(rw)
	//}

	//{
	//	doLogin(rw)
	//	doCloseBinding(rw)
	//}

	//{
	//	doLogin(rw)
	//	getApplyUsers(rw)
	//}

	//{
	//	doLogin(rw)
	//	approvalApply(rw, "14fzqramdntchdt8qz6ka93ehprjryeb7nf2cj9hp5x03dtknr5skdjka", 1, "son")
	//}

	//{
	//	doLogin(rw)
	//	getShareUsers(rw)
	//}

	loggermsg.Info("working done~~~~~~~~~~~~")
}
