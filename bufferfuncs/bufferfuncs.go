package bufferfuncs

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"strings"
)

func ReadVarInt(buf []byte, currentByteOffset int) (byteOffset int, value uint) {
	byteOffset = VarIntLength(buf[currentByteOffset:])
	value = VarIntValue(buf[currentByteOffset:], byteOffset)

	return
}

// convenience function for returning the
// byte offset of a var_int and the value it represents.

// As per bitcoin p2p protocol, checks the first byte
// of a var_int for a leading byte
// connoting a large unsigned integer value.
func VarIntLength(buf []byte) (byteOffset int) {
	firstByte := hex.EncodeToString(buf[0:1])
	switch firstByte {
	case "fd": // >= 253, uint_16
		byteOffset = 3
	case "fe": // > 65535, uint_32
		byteOffset = 5
	case "ff": // > 4294967295, uint_64
		byteOffset = 9
	default: // uint_8
		byteOffset = 1
	}

	return
}

// Sadness
func VarIntValue(buf []byte, byteOffset int) (val uint) {
	switch byteOffset {
	case 1:
		data, err := ReadUInt8(buf[0:1])
		if err != nil {
			log.Error(err)
		}
		val = uint(data)
	case 3:
		data, err := ReadUInt16(buf[1:3])
		if err != nil {
			log.Error(err)
		}
		val = uint(data)
	case 5:
		data, err := ReadUInt32(buf[1:5])
		if err != nil {
			log.Error(err)
		}
		val = uint(data)
	case 9:
		data, err := ReadUInt64(buf[1:9])
		if err != nil {
			log.Error(err)
		}
		val = uint(data)
	}

	return
}

// Hacky pointer magic.
// Reads little-endian data from a byte slice into dataPtr,
// which will always be a pointer to some kind of int8/16/etc
// or uint8/16/etc.
func readIntFromBuffer(buf []byte, dataPtr interface{}) (err error) {
	buffer := bytes.NewBuffer(buf)
	err = binary.Read(buffer, binary.LittleEndian, dataPtr)

	return
}

// Below this be wrapper functions.

func ReadInt32(buf []byte) (ret int32, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadInt64(buf []byte) (ret int64, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadUInt8(buf []byte) (ret uint8, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadUInt16(buf []byte) (ret uint16, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadUInt32(buf []byte) (ret uint32, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadUInt64(buf []byte) (ret uint64, err error) {
	err = readIntFromBuffer(buf, &ret)

	return
}

func ReadString(buf []byte) (fieldVal string) {
	fieldVal = string(buf)

	return
}

func ReverseStringSlice(s []string) []string {
	for i := len(s)/2 - 1; i >= 0; i-- {
		opp := len(s) - 1 - i
		s[i], s[opp] = s[opp], s[i]
	}

	return s
}

// Takes a []byte which represents a string in
// little-endian format, and returns said string
// in big-endian format.
func GetBigEndianHexString(buf []byte) (hexHash string) {
	var hexList []string
	reversedHash := hex.EncodeToString(buf)

	for i := 0; i < len(reversedHash); i += 2 {
		hexList = append(hexList, reversedHash[i:i+2])
	}

	hexList = ReverseStringSlice(hexList)

	hexHash = strings.Join(hexList, "")

	return
}
