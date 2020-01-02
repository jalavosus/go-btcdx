package main

import (
	"fmt"

	"github.com/jalavosus/go-btcdx/btcpeer"
)

func main() {
	peer := btcpeer.NewPeer("54.212.120.67", 8333)

	fmt.Println("Got peer")

	peer.WaitForDisconnect()
}
