package bbcsign

import (
	"encoding/hex"
	"gitlab.dabank.io/nas/app-simulator/utils"
	"testing"
)

func TestByte2Addr(t *testing.T) {
	pukBuf, _ := hex.DecodeString("8716ef8af7dd5017652a2e5b79aaa99ad7bd83fd7921c024169c7bdadd839286")
	pukBuf = utils.ReverseBytes(pukBuf)
	pukBuf = append([]byte{0x01}, pukBuf...)
	addr := PukByte2Addr(pukBuf)
	t.Log("addr:", addr)
	if addr != "1gt987qetfee1c96045wzv0xxtydakaksbcq2ms8qa3ezf2qf2t3zy534" {
		t.Fatal("invalid wallet address")
	}
}
