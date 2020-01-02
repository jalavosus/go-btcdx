package message

import (
	"bytes"

	"github.com/btcsuite/btcd/wire"
	bf "github.com/jalavosus/go-btcdx/bufferfuncs"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
}

type Message interface {
	GetPayload() interface{}
}

// Header data for a received bitcoin p2p message.
type MessageHeader struct {
	NetworkMagic    uint32 `json:"magic"`    // Magic integer value connoting the network this message came from.
	Command         string `json:"cmd"`      // Command that was called by the receiver of this message.
	PayloadLen      uint32 `json:"len"`      // Length of the payload in bytes.
	PayloadChecksum uint32 `json:"checksum"` // First 4 bytes of sha256(sha256(payload)).
}

// Reads the first 24 bytes of an incoming btcd peer message,
// reading its header info.
func ParseHeader(msgBytes []byte) (header MessageHeader) {
	magic, err := bf.ReadUInt32(msgBytes[0:4])
	checkError(err)

	cmd := bf.ReadString(msgBytes[4:16])
	payloadLen, err := bf.ReadUInt32(msgBytes[16:20])
	checkError(err)

	checksum, err := bf.ReadUInt32(msgBytes[20:24])
	checkError(err)

	header = MessageHeader{
		NetworkMagic:    magic,
		Command:         cmd,
		PayloadLen:      payloadLen,
		PayloadChecksum: checksum,
	}

	return
}

// Reads data from an incoming message, returning both witness-encoded and
// non-witness-encoded byte slices.
// Panics if an incoming message has fewer than 24 bytes (no header) or if
// any other kind of error occurs.
func ReadMessageBytes(msg wire.Message) (nonWitnessBytes []byte, witnessBytes []byte) {
	var (
		nonWitBuf, witBuf *bytes.Buffer
	)

	nonWitBuf = new(bytes.Buffer)
	witBuf = new(bytes.Buffer)

	bytesRead, err := wire.WriteMessageN(nonWitBuf, msg, 2, wire.MainNet)
	checkBytesRead(bytesRead)
	checkError(err)

	bytesRead, err = wire.WriteMessageWithEncodingN(witBuf, msg, 2, wire.MainNet, wire.WitnessEncoding)
	checkBytesRead(bytesRead)
	checkError(err)

	nonWitnessBytes = nonWitBuf.Bytes()
	witnessBytes = witBuf.Bytes()

	return
}

func checkBytesRead(bytesRead int) {
	if bytesRead < 24 {
		log.Panic("Fewer than 24 bytes read from incoming message, this means there is no message header.")
	}
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
