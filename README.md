# go-btcdx 

[btcsuite/btcd](https://github.com/btcsuite/btcd/) is a fantastic library. 

The data it exposes might now be enough, however. This is where **go-btcdx** comes in.

By providing drop-in functions for Transactions, Blocks, etc., **go-btcdx** enables
reading ALL of the incoming data from a btcd peer connection. 

## Installation (with gomodules)

`import "github.com/jalavosus/go-btcdx"`

## Usage (Example)

When configuring message listeners in your [btcd.Peer](https://github.com/btcsuite/btcd/tree/master/peer) configuration,
instead of using `MessageListeners.OnTx` (for example), extend `MessageListeners.OnRead`, like so:

```go
import (
	"fmt"
	btcdtx "github.com/jalavosus/go-btcdx/tx"
)

var peerConfig *peer.Config = &peer.Config{
	UserAgentName: "", // User agent name to advertise.
	ChainParams: &chaincfg.MainNetParams,
	DisableRelayTx: false,
	Listeners: peer.MessageListeners{
		OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
			return nil
		},
		OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
			doneChan := make(chan struct{})
			newMsg := wire.NewMsgGetData()
			for _, invVector := range msg.InvList {
				invVector.Type = wire.InvTypeWitnessTx
				if err := newMsg.AddInvVect(invVector); err != nil {
					log.Error(err)
				}
			}
			p.QueueMessage(newMsg, doneChan)
			<-doneChan
		},
		OnRead: func(p *peer.Peer, bytesRead int, msg wire.Message, err error) {
			if err != nil {
				log.Error(err)
				return
			}
			if msg.Command() == "tx" {
				newTx := btcdtx.DecodeMessage(p, bytesRead, msg, err)
				fmt.Println(newTx.Payload.Hash)
			}
		},
	},
}
```