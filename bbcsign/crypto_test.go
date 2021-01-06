package bbcsign

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"gitlab.dabank.io/nas/app-simulator/utils"
	"testing"
)

func TestVerify(t *testing.T) {
	data := []byte("security module sign test")
	signature, _ := hex.DecodeString("fd9dfd3e42e67d70c55f132b2261fa09c7f4552cea507c29f75dff877628d250879e1a282b052ae8d9bf63f8691d8a896ef0b402bd95ff63d5e56531815e5e05")
	pass := Verify(data, signature)
	if !pass {
		t.Error("verify fail")
	}
}

func TestCreatePublicKey(t *testing.T) {
	key, err := hex.DecodeString("9d6f2ee97ffefb292581496a3449d68bf9f15c3ecbfb841b4683b1772e72115b")
	if err != nil {
		t.Error(err)
	}

	key = utils.ReverseBytes(key)
	sk = ed25519.NewKeyFromSeed(key)
	fmt.Println("public key:", hex.EncodeToString(sk.Public().(ed25519.PublicKey)))
}
