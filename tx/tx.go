package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	bf "github.com/jalavosus/go-btcdx/bufferfuncs"
	"github.com/jalavosus/go-btcdx/message"
	log "github.com/sirupsen/logrus"
)

const (
	OutpointSize     int    = 36
	HeaderSize       int    = 24
	LocktimeBoundary uint32 = 500000000
)

func DecodeMessage(p *peer.Peer, bytesRead int, msg wire.Message, err error) (decodedMsg TransactionMessage) {
	timestamp := time.Now().Unix()

	nonWitnessBytes, witnessBytes := message.ReadMessageBytes(msg)

	witnessHeaderBytes := witnessBytes[0:HeaderSize]

	msgHeader := message.ParseHeader(witnessHeaderBytes)

	witnessPayloadBytes := witnessBytes[HeaderSize:]

	msgPayload := parsePayload(witnessPayloadBytes, timestamp)

	hash := parseTxid(nonWitnessBytes[24:])
	witnessHash := parseTxid(witnessBytes[24:])

	msgPayload.Hash = hash
	msgPayload.WitnessHash = witnessHash

	decodedMsg.Header = msgHeader
	decodedMsg.Payload = msgPayload

	return
}

func parsePayload(buf []byte, timestamp int64) TX {
	var (
		inTxns            []TXIn
		outTxns           []TXOut
		currentByteOffset int // implicit 0
		hasWitnessFlag    bool
		flag              uint8
	)

	txVersion, _ := bf.ReadInt32(buf[currentByteOffset:4])

	currentByteOffset += 4

	hasWitnessFlag, flag = getWitnessFlag(buf[currentByteOffset : currentByteOffset+2])

	if hasWitnessFlag {
		currentByteOffset += 2
	}

	inputOffset, inputCount := bf.ReadVarInt(buf, currentByteOffset)

	currentByteOffset += inputOffset

	locktime, _ := bf.ReadUInt32(buf[len(buf)-4:])

	buf = buf[0 : len(buf)-4] // Strip out locktime bytes from the buffer to make parsing easier.

	currentByteOffset, inTxns = parseInputs(buf, currentByteOffset, inputCount)

	outputOffset, outputCount := bf.ReadVarInt(buf, currentByteOffset)

	currentByteOffset += outputOffset

	currentByteOffset, outTxns = parseOutputs(buf, currentByteOffset, outputCount)

	newTx := TX{
		Version:      txVersion,
		Locktime:     locktime,
		LocktimeType: parseLocktype(locktime),
		InCount:      inputCount,
		InTxns:       inTxns,
		OutCount:     outputCount,
		OutTxns:      outTxns,
		Timestamp:    timestamp,
	}

	if hasWitnessFlag {
		newTx.Flag = &flag
		_, witnesses := parseWitnessData(buf, currentByteOffset, int(inputCount))
		newTx.Witnesses = &witnesses
	}

	return newTx
}

// Returns a transactions txid.
// txid is defined as sha256(sha256(rawTxn))
func parseTxid(buf []byte) string {
	innerSum := sha256.Sum256(buf)
	realInnerSum := []byte(innerSum[:]) // sha256.Sum256 returns a [32]byte, which is incompatible with []byte
	outerSum := sha256.Sum256(realInnerSum)
	realOuterSum := []byte(outerSum[:])

	txid := bf.GetBigEndianHexString(realOuterSum)

	return txid
}

func getWitnessFlag(buf []byte) (hasWitnessFlag bool, flag uint8) {
	marker, err := bf.ReadUInt8(buf[0:1])
	if err != nil {
		log.Error(err)
	}
	flag, err = bf.ReadUInt8(buf[1:])
	if err != nil {
		log.Error(err)
	}

	hasWitnessFlag = marker == 0 && flag == 1

	return
}

// Returns all of a transactions inputs.
// bytesRead is the new byte offset for other functions to use.
func parseInputs(buf []byte, currentByteOffset int, inputCount uint) (bytesRead int, inTxns []TXIn) {
	for i := 0; i < int(inputCount); i++ {
		var (
			script    string
			scriptLen int
			txIn      TXIn
		)

		outpointData := buf[currentByteOffset : currentByteOffset+OutpointSize]
		txIn.PreviousOutput = parseOutpoint(outpointData)

		currentByteOffset += OutpointSize

		currentByteOffset, scriptLen, script = parseScript(buf, currentByteOffset)

		txIn.Script = script
		txIn.ScriptLen = scriptLen

		sequence, _ := bf.ReadUInt32(buf[currentByteOffset : currentByteOffset+4])
		txIn.Sequence = sequence

		currentByteOffset += 4

		inTxns = append(inTxns, txIn)
	}

	bytesRead = currentByteOffset

	return
}

// Returns all of a transactions outputs.
// bytesRead is the new byte offset for other functions to use.
func parseOutputs(buf []byte, currentByteOffset int, outputCount uint) (bytesRead int, outTxns []TXOut) {
	for i := 0; i < int(outputCount); i++ {
		var (
			value     int64
			script    string
			scriptLen int
			outTx     TXOut
		)

		value, _ = bf.ReadInt64(buf[currentByteOffset : currentByteOffset+8])
		currentByteOffset += 8

		currentByteOffset, scriptLen, script = parseScript(buf, currentByteOffset)

		outTx.Value = value
		outTx.PKScript = script
		outTx.PKScriptLen = scriptLen

		outTxns = append(outTxns, outTx)
	}

	bytesRead = currentByteOffset

	return
}

func parseWitnessData(buf []byte, currentByteOffset int, numInputs int) (bytesRead int, txWitnesses []TXWitness) {
	for i := 0; i < numInputs; i++ {
		var (
			witness    TXWitness
			components []string
		)

		offset, numComponents := bf.ReadVarInt(buf, currentByteOffset)
		currentByteOffset += offset

		witness.Count = int(numComponents)

		// loop through witness components
		for j := 0; j < int(numComponents); j++ {
			compOffset, compLen := bf.ReadVarInt(buf, currentByteOffset)
			currentByteOffset += compOffset

			var component string

			// There can actually be 0-len witness components. Kinda weird, but it happens.
			if compLen > 0 {
				component = hex.EncodeToString(buf[currentByteOffset : currentByteOffset+int(compLen)])
			}

			components = append(components, component)

			currentByteOffset += int(compLen)
		}

		witness.Witnesses = components

		txWitnesses = append(txWitnesses, witness)
	}

	bytesRead = currentByteOffset

	return
}

// Returns an input's Outpoint information.
func parseOutpoint(buf []byte) (op Outpoint) {
	var (
		hash  string
		index uint32
	)

	hashBin := buf[0:32]
	hash = bf.GetBigEndianHexString(hashBin)

	index, _ = bf.ReadUInt32(buf[32:])

	op.Hash = hash
	op.Index = index

	return
}

// Returns scripts for an input or output.
// bytesRead is the new byte offset for other functions to use.
func parseScript(buf []byte, currentByteOffset int) (bytesRead int, scriptLen int, script string) {
	byteOffset, scriptLn := bf.ReadVarInt(buf, currentByteOffset)
	scriptLen = int(scriptLn)

	bufStart := currentByteOffset + byteOffset
	scriptBuf := buf[bufStart : bufStart+scriptLen]

	scriptOffset := byteOffset + scriptLen
	currentByteOffset += scriptOffset // Increment offset by number of bytes for size and script size in bytes

	if scriptLen > 0 {
		script = hex.EncodeToString(scriptBuf)
	} else {
		script = ""
	}

	bytesRead = currentByteOffset

	return
}

func parseLocktype(locktime uint32) (locktype string) {
	switch {
	case locktime == 0:
		locktype = "not_locked"
	case locktime < LocktimeBoundary:
		locktype = "block"
	case locktime >= LocktimeBoundary:
		locktype = "timestamp"
	}

	return
}
