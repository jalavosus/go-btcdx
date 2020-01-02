package tx

import (
	. "github.com/jalavosus/go-btcdx/message"
)

// Generic message container. Always contains a header,
// almost always contains a payload.
type TransactionMessage struct {
	Header  MessageHeader `json:"header"`
	Payload TX            `json:"payload"`
}

func (t *TransactionMessage) GetPayload() interface{} {
	return t.Payload
}

// TX payload structure.
type TX struct {
	Version      int32        `json:"version"`             // TX data format version. Signed. 4 bytes
	Flag         *uint8       `json:"flag,omitempty"`      // Witness data existence flag. Optional. 0 or 2 bytes.
	InCount      uint         `json:"in_count"`            // Number of input transactions. Count will never be 0. >= 1 bytes.
	InTxns       []TXIn       `json:"in_txns"`             // Array of input transactions. >= 41 bytes.
	OutCount     uint         `json:"out_count"`           // Number of output transactions. Count >= 1. >= 1 bytes
	OutTxns      []TXOut      `json:"out_txns"`            // Array of output transactions. >= 41 bytes.
	Witnesses    *[]TXWitness `json:"witnesses,omitempty"` // If Flag is present, array of witnesses (1 per input). Omitted if Flag is omitted. >= 0 bytes.
	Locktime     uint32       `json:"locktime"`            // Block number or timestamp at which this transaction is unlocked.4 bytes.
	Hash         string       `json:"hash"`                // sha256-computed txid, non-witness
	WitnessHash  string       `json:"witness_hash"`        // sha256-computer txid, witness
	LocktimeType string       `json:"locktime_type"`       // Whether the Tx's locktime is unlocked, a block, or a unix timestamp
	Timestamp    int64        `json:"timestamp"`           // Generated when the peer receives tx data.
}

type TXIn struct {
	PreviousOutput Outpoint           `json:"prev_output"` // Reference to the previous transaction output. 36 bytes.
	ScriptLen      int                `json:"script_len"`  // Length of the signature script in bytes. >= 1 bytes.
	Script         string             `json:"script"`      // Computational Script for confirming transaction authorization. ScriptLen bytes.
	Sequence       uint32             `json:"sequence"`    // Transaction version as defined by the sender.
	Witnesses      []WitnessComponent `json:"witnesses"`   // TXIn witnesses.
}

type TXOut struct {
	Value       int64  `json:"value"`         // Transaction value. 8 bytes.
	PKScriptLen int    `json:"pk_script_len"` // Length of the output's public key script. >= 1 bytes.
	PKScript    string `json:"pk_script"`     // Public key script. PKScriptLen bytes.
}

type TXWitness struct {
	Count     int      `json:"count"`     // Number of witness components. >= 1 bytes.
	Witnesses []string `json:"witnesses"` // Actual array of witness components. ? bytes.
}

type WitnessComponent struct {
	Length int    `json:"length"` // Length of witness component data. >= 1 bytes.
	Data   string `json:"data"`   // Raw witness component data. Length bytes.
}

type Outpoint struct {
	Hash  string `json:"hash"`  // Hash of the referenced output transaction. Char len 32. 32 bytes.
	Index uint32 `json:"index"` // Index of the specific output in the transaction. 4 bytes.
}
