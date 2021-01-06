package bbcsign

import (
	"crypto/ed25519"
	"encoding/hex"
	"gitlab.dabank.io/nas/app-simulator/utils"
	"golang.org/x/crypto/blake2b"
)

type KeyInfo struct {
	PriKey []byte
	PubKey []byte
	Addr   string
}

func Sign(key string, data []byte) []byte {
	k, _ := hex.DecodeString(key)
	k = utils.ReverseBytes(k)
	sk := ed25519.NewKeyFromSeed(k)
	hash := blake2b.Sum256(data)
	return ed25519.Sign(sk, hash[:])
}

func Verify(data, signature []byte) bool {
	//hash := blake2b.Sum256(data)
	//
	//edpub, ok := sk.Public().(ed25519.PublicKey)
	//if !ok {
	//	return false
	//}
	//
	//return ed25519.Verify(edpub, hash[:], signature)
	return true
}

func VerifyWithAddr(addr string, data, signature []byte) bool {
	pukBuf := Addr2PukByte(addr)
	if len(pukBuf) != 32 {
		return false
	}

	edPub := ed25519.PublicKey(pukBuf)
	return ed25519.Verify(edPub, data, signature)
}

func Ed25519Sign(key string, data []byte) []byte {
	k, _ := hex.DecodeString(key)
	k = utils.ReverseBytes(k)
	sk := ed25519.NewKeyFromSeed(k)
	return ed25519.Sign(sk, data)
}

func Ed25519Verify(puk string, data, sig []byte) bool {
	pukb, _ := hex.DecodeString(puk)
	pukb = utils.ReverseBytes(pukb)
	edpub := ed25519.PublicKey(pukb)
	return ed25519.Verify(edpub, data, sig)
}

func PrintPukToAddress(puk []byte) string {
	rpk := utils.ReverseBytes(puk)
	rpk = append([]byte{0x01}, rpk...)
	return PukByte2Addr(rpk)
}

func GetFullInfoFromPrintKey(key []byte) *KeyInfo {
	var info KeyInfo
	info.PriKey = key
	rKey := utils.ReverseBytes(key)
	sk := ed25519.NewKeyFromSeed(rKey)
	edPub, _ := sk.Public().(ed25519.PublicKey)
	info.PubKey = utils.ReverseBytes(edPub)

	edPub = append([]byte{0x01}, edPub...)
	info.Addr = PukByte2Addr(edPub)
	return &info
}
