package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"gitlab.dabank.io/nas/go-msgbase/p2pprotocol"
	"gitlab.dabank.io/nas/go-msgbase/udpprotocol"
	. "gitlab.dabank.io/nas/go-msgbase/saferw"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Binding struct {
	devices    map[string]net.Addr
	mux        sync.Mutex
	stopListen chan struct{}
	pc         net.PacketConn
}

var (
	LISTEN_HELLO_ADDR = ":52030"
	COMMUNICATE_ADDR  = ":52031"
)

func getBindingStatus(rw *SafeRW) {
	loggermsg.Info("isBinding")

	var req p2pprotocol.GetBindingStatus
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.GET_BINDINGSTATUS, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.GET_BINDINGSTATUS_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.GET_BINDINGSTATUS_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.GetBindingStatusResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal GetBindingStatusResp fail, err:", err)
		return
	}

	loggermsg.Info("GetBindingStatusResp:", resp)
}

func doOpenBinding(rw *SafeRW) {
	loggermsg.Info("doOpenBinding")

	var req p2pprotocol.OpenBinding
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.OPEN_BINDING, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.OPEN_BINDING_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.OPEN_BINDING_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.OpenBindingResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal OpenBindingResp fail, err:", err)
		return
	}

	loggermsg.Info("OpenBindingResp:", resp)
}


func doCloseBinding(rw *SafeRW) {
	loggermsg.Info("doCloseBinding")

	var req p2pprotocol.CloseBinding
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.CLOSE_BINDING, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.CLOSE_BINDING_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.CLOSE_BINDING_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.CloseBindingResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal CloseBindingResp fail, err:", err)
		return
	}

	loggermsg.Info("CloseBindingResp:", resp)
}

func getApplyUsers(rw *SafeRW) {
	loggermsg.Info("getApplyUsers")

	var req p2pprotocol.GetApplyUsers
	req.Nonce = rand.Uint32()

	sendMsg(p2pprotocol.GET_APPLYUSERS, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.GET_APPLYUSERS_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.GET_APPLYUSERS_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.GetApplyUsersResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal GetApplyUsersResp fail, err:", err)
		return
	}

	loggermsg.Info("GetApplyUsersResp:", resp)
}


func approvalApply(rw *SafeRW, addr string, pass uint32, nickname string) {
	loggermsg.Info("approvalApply")

	var req p2pprotocol.ApprovalApply
	req.Nonce = rand.Uint32()
	req.BbcAddr = addr
	req.Pass = pass
	req.Nickname = nickname

	sendMsg(p2pprotocol.APPROVAL_APPLY, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.APPROVAL_APPLY_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.APPROVAL_APPLY_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.ApprovalApplyResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal ApprovalApplyResp fail, err:", err)
		return
	}

	loggermsg.Info("ApprovalApplyResp:", resp)
}

func getShareUsers(rw *SafeRW) {
	loggermsg.Info("getShareUsers")

	var req p2pprotocol.GetShareUsers
	req.Nonce = rand.Uint32()
	sendMsg(p2pprotocol.GET_SHAREUSERS, &req, rw)

	body := make([]byte, 1024)
	msglen, cmdid, body, err := ReadOneMsg(rw, body)
	if err != nil {
		loggermsg.Error("read resp msg fail. err:", err)
		return
	}

	if p2pprotocol.P2pMsgID(cmdid) != p2pprotocol.GET_SHAREUSERS_RESP {
		loggermsg.Error("invalid resp msg cmd id, expect:", p2pprotocol.GET_SHAREUSERS_RESP, ", actual:", cmdid)
	}

	var resp p2pprotocol.GetShareUsersResp
	err = proto.Unmarshal(body[0:msglen-8], &resp)
	if err != nil {
		loggermsg.Error("proto unmarshal GetShareUsersResp fail, err:", err)
		return
	}

	loggermsg.Info("GetShareUsersResp:", resp)
}


func bind() {
	var bind Binding
	bind.Init()
	bind.StartListenNas()
	var deviceid string
	for {
		var list []string
		jsonstr, _ := bind.GetDeviceList()
		fmt.Println("ids:", jsonstr)
		json.Unmarshal([]byte(jsonstr), &list)
		if len(list) == 0 {
			time.Sleep(time.Second)
			continue
		}

		deviceid = list[0]
		break
	}

	fmt.Println("deviceid:", deviceid)
	key, _ := hex.DecodeString("aabb112233445566778899001122334455667788990011223344556677889900")
	nasid, err := bind.Bind(deviceid, key, 10)
	if err != nil {
		loggermsg.Error("bind fail, err:", err)
		return
	}

	time.Sleep(time.Second*2)
	bind.Bye(deviceid, 5)

	loggermsg.Info("nas id:", nasid)
	return
}

// 初始化Binding对象
func (b *Binding) Init() error {
	b.devices = make(map[string]net.Addr)
	b.stopListen = make(chan struct{}, 2)
	return nil
}

// 启动监听Nas设备广播
func (b *Binding) StartListenNas() error {
	loggermsg.Info("start listen nas udp msg")
	pc, err := net.ListenPacket("udp4", LISTEN_HELLO_ADDR)
	if err != nil {
		loggermsg.Error("listen udp packet fail", "addr", LISTEN_HELLO_ADDR, "err", err)
		return err
	}

	go b.startListen(pc)
	return nil
}

// 停止监听Nas设备
func (b *Binding) StopListenNas() error {
	// todo: 停止监听Nas设备，释放占用资源
	return nil
}

// 返回已经监听得到的deviceId，
// 返回值：
//		string：[]string的json序列化字符串，存放所有设备的deviceId
func (b *Binding) GetDeviceList() (string, error) {
	b.mux.Lock()
	list := make([]string, 0, len(b.devices))
	for k, _ := range b.devices {
		list = append(list, k)
	}
	defer b.mux.Unlock()
	sort.Strings(list)
	ret, err := json.Marshal(&list)
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

// 绑定指定设备
// 参数：
//		deviceid：设备id
//		key：钱包私钥
//		timeout：超时时间
// 返回：
//		string：nas id（nas设备激活id）
func (b *Binding) Bind(deviceid string, key []byte, timeout int) (string, error) {
	if len(key) != 32 {
		return "", errors.New("invalid key len")
	}

	b.mux.Lock()
	v, ok := b.devices[deviceid]
	defer b.mux.Unlock()
	if !ok {
		loggermsg.Error("invalid deviceid to bind")
		return "", errors.New("invalid deviceid")
	}

	pc, err := net.ListenPacket("udp4", COMMUNICATE_ADDR)
	if err != nil {
		loggermsg.Error("listen udp packet fail", "addr", COMMUNICATE_ADDR, "err", err)
		return "", err
	}
	defer pc.Close()

	bg := time.Now()
	for {
		if time.Since(bg) > time.Second*time.Duration(timeout) {
			return "", errors.New("binding timeout")
		}

		id, err := bindProcess(deviceid, pc, v, key, timeout)
		if err != nil {
			loggermsg.Warn("binding fail, wait a moment. deviceid:", deviceid)
			time.Sleep(time.Second)
			continue
		}

		loggermsg.Debug("binding success. deviceid:", deviceid)
		return id, nil
	}
}

// 通知指定设备绑定成功
// 参数：
//		deviceid：nas设备id
//		timeout：超时时间
func (b *Binding) Bye(deviceid string, timeout int) error {
	b.mux.Lock()
	raddr, ok := b.devices[deviceid]
	b.mux.Unlock()
	if !ok {
		return errors.New("invalid deviceid")
	}

	pc, err := net.ListenPacket("udp4", COMMUNICATE_ADDR)
	if err != nil {
		loggermsg.Error("listen udp packet fail", "addr", COMMUNICATE_ADDR, "err", err)
		return err
	}
	defer pc.Close()

	bg := time.Now()
	for {
		if time.Since(bg) > time.Second*time.Duration(timeout) {
			loggermsg.Error("bye timeout")
			return errors.New("bye timeout")
		}

		err = sendBey(pc, raddr)
		if err != nil {
			continue
		}

		err = readBey(pc, raddr, 3)
		if err != nil {
			continue
		}

		for i := 0; i < 5; i++ {
			sendBey(pc, raddr)
			time.Sleep(time.Millisecond * 100)
		}

		return nil
	}

	return nil
}

func (b *Binding) startListen(pc net.PacketConn) {
	buf := make([]byte, 1024)
	for {
		hello, raddr, err := waitHello(pc, buf)
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}

		b.mux.Lock()
		b.devices[hello.PeerId] = raddr
		b.mux.Unlock()
	}
}

func bindProcess(deviceid string, pc net.PacketConn, raddr net.Addr, key []byte, timeout int) (string, error) {
	err := sendKey(deviceid, pc, raddr, key)
	if err != nil {
		return "", err
	}

	bg := time.Now()
	rbuf := make([]byte, 1024*256)
	for {
		if time.Since(bg) > time.Second*time.Duration(timeout) {
			return "", errors.New("timeout")
		}

		pc.SetReadDeadline(time.Now().Add(time.Second * 2))
		n, raddrnew, err := pc.ReadFrom(rbuf)
		if err != nil {
			loggermsg.Error("read msg from box fail", "err", err)
			return "", err
		}

		if n < 3 || raddrnew.String() != raddr.String() {
			continue
		}

		if udpprotocol.UdpMsgID(rbuf[0]) != udpprotocol.UDP_NASID {
			loggermsg.Error("invalid received udp msg id", "expect", udpprotocol.UDP_NASID, "acture", rbuf[0])
			return "", errors.New("invalid box resp msg")
		}

		if !udpprotocol.CheckCrc16(rbuf[:n]) {
			loggermsg.Error("invalid crc16 signature")
			return "", errors.New("invalid crc16 signature")
		}

		return "", nil
	}

}

func sendBey(pc net.PacketConn, raddr net.Addr) error {
	loggermsg.Debug("send bye begin")
	defer loggermsg.Debug("send bye end")

	var bye udpprotocol.UdpBye
	bye.Timestamp = uint64(time.Now().Unix())
	buf, err := proto.Marshal(&bye)
	if err != nil {
		loggermsg.Error("protobuf marshal bye object fail", "err", err)
		return err
	}

	msgBuflen := len(buf) + 3
	msgBuf := make([]byte, msgBuflen)
	msgBuf[0] = byte(udpprotocol.UDP_BYE)
	copy(msgBuf[1:], buf)
	udpprotocol.WriteCrc16(msgBuf)
	pc.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, err = pc.WriteTo(msgBuf, raddr)
	if err != nil {
		loggermsg.Error("send udp package fail", "err", err)
		return err
	}

	return nil
}

func readBey(pc net.PacketConn, raddr net.Addr, timeout int) error {
	rbuf := make([]byte, 1024*256)
	bg := time.Now()
	for {
		if time.Since(bg) > time.Second*time.Duration(timeout) {
			return errors.New("read bye timeout")
		}

		pc.SetReadDeadline(time.Now().Add(time.Second * 2))
		n, raddrnew, err := pc.ReadFrom(rbuf)
		if err != nil {
			loggermsg.Error("read msg from box fail", "err", err)
			return err
		}

		if n < 3 || raddrnew.String() != raddr.String() {
			continue
		}

		if udpprotocol.UdpMsgID(rbuf[0]) != udpprotocol.UDP_BYE {
			return errors.New("invalid box resp msg")
		}

		if !udpprotocol.CheckCrc16(rbuf[:n]) {
			loggermsg.Error("invalid crc16 signature")
			return errors.New("invalid crc16 signature")
		}

		var msg udpprotocol.UdpBye
		err = proto.Unmarshal(rbuf[1:n-2], &msg)
		if err != nil {
			loggermsg.Error("protobuf unmarshal bye object fail. err:", err)
			return err
		}

		return nil
	}
}

func sendKey(deviceid string, pc net.PacketConn, raddr net.Addr, key []byte) error {
	loggermsg.Debug("send key begin")
	defer loggermsg.Debug("send key end")

	s := sha256.Sum256([]byte(deviceid))
	cipherKey, err := udpprotocol.AES256Encrypt(s[:], key)
	if err != nil {
		return err
	}

	var privatekey udpprotocol.UdpPrivateKey
	privatekey.Timestamp = uint64(time.Now().Unix())
	privatekey.Key = cipherKey
	m := md5.New()
	m.Write([]byte(strconv.FormatInt(int64(privatekey.Timestamp), 10)))
	m.Write(cipherKey)
	privatekey.Sign = m.Sum(nil)
	buf, err := proto.Marshal(&privatekey)
	if err != nil {
		loggermsg.Error("protobuf marshal private key object fail", "err", err)
		return err
	}

	msgBuflen := len(buf) + 3
	msgBuf := make([]byte, msgBuflen)
	msgBuf[0] = byte(udpprotocol.UDP_PRIVATEKEY)
	copy(msgBuf[1:], buf)
	udpprotocol.WriteCrc16(msgBuf)
	pc.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, err = pc.WriteTo(msgBuf, raddr)
	if err != nil {
		loggermsg.Error("send udp package fail", "err", err)
		return err
	}

	return nil
}

func waitHello(pc net.PacketConn, buf []byte) (*udpprotocol.UdpHello, net.Addr, error) {
	for {
		msgId, body, raddr, err := waitMsg(pc, buf)
		if err != nil {
			return nil, nil, err
		}

		if msgId != udpprotocol.UDP_HELLO {
			continue
		}

		var msg udpprotocol.UdpHello
		err = proto.Unmarshal(body, &msg)
		if err != nil {
			continue
		}

		if !verifyHello(&msg) {
			loggermsg.Warn("hello msg verify fail")
			continue
		}

		loggermsg.Info("get a hello msg, PeerId:", msg.PeerId)
		return &msg, raddr, nil
	}
}

func waitMsg(pc net.PacketConn, buf []byte) (udpprotocol.UdpMsgID, []byte, net.Addr, error) {
	for {
		loggermsg.Info("before read a udp msg")
		pc.SetReadDeadline(time.Now().Add(time.Second * 2))
		n, raddr, err := pc.ReadFrom(buf)
		loggermsg.Info("read a udp msg, ", "n:", n)
		if n < 3 {
			continue
		} else if err != nil && errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		} else if err != nil {
			return 0, nil, nil, err
		}

		if !udpprotocol.CheckCrc16(buf[:n]) {
			loggermsg.Warn("crc16 verify fail")
			continue
		}

		return udpprotocol.UdpMsgID(buf[0]), buf[1 : n-2], raddr, nil
	}
}

func verifyHello(hello *udpprotocol.UdpHello) bool {
	m := md5.New()
	m.Write([]byte(strconv.FormatInt(int64(hello.Timestamp), 10)))
	m.Write([]byte(hello.ClientName))
	m.Write([]byte(hello.PeerId))
	return bytes.Compare(hello.Sign, m.Sum(nil)) == 0
}

func startUDPClient() {
	pc, err := net.ListenPacket("udp4", ":52021")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	addr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:52022")
	if err != nil {
		panic(err)
	}

	i := 0
	readBuf := make([]byte, 1024)
	for {
		i++
		_, err = pc.WriteTo([]byte("data to transmit "+strconv.Itoa(i)), addr)
		if err != nil {
			panic(err)
		}

		n, raddr, err := pc.ReadFrom(readBuf)
		if err != nil {
			fmt.Println("read package fail, err:", err)
			return
		}

		fmt.Printf("%s sent this: %s\n", raddr, readBuf[:n])
		time.Sleep(time.Second)
	}
}

func startUDPServer() {
	pc, err := net.ListenPacket("udp4", ":52022")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	buf := make([]byte, 1024)
	i := 0
	for {
		n, raddr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s sent this: %s\n", raddr, buf[:n])

		i++
		pc.WriteTo([]byte("resp "+strconv.Itoa(i)), raddr)
	}
}
