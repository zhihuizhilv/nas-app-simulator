module gitlab.dabank.io/nas/app-simulator

go 1.14

require (
	github.com/golang/protobuf v1.4.2
	github.com/ipfs/go-log/v2 v2.1.1
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/snksoft/crc v1.1.0
	gitlab.dabank.io/nas/go-nas v1.0.0
	gitlab.dabank.io/nas/p2p-network v1.0.5
	google.golang.org/protobuf v1.23.0

)

// replace gitlab.dabank.io/nas/p2p-network => /home/lzh/code/gopath/src/gitlab.dabank.io/nas/p2p-network

replace gitlab.dabank.io/nas/go-nas => /home/lzh/code/gopath/src/gitlab.dabank.io/nas/go-nas
