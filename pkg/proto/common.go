package proto

import (
	"encoding/binary"
)

func StringFromBytes(data []byte) (string, []byte) {
	length, data := IntFromBytes(data)
	utfString := string(data[:length])
	return utfString, data[length:]
}

func StringToBytes(value string) []byte {
	rawString := []byte(value)
	length := uint32(len(rawString))
	rawLength := IntToBytes(length)
	return append(rawLength, rawString...)
}

func IntFromBytes(data []byte) (uint32, []byte) {
	value := binary.BigEndian.Uint32(data)
	return value, data[4:]
}

func IntToBytes(value uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)
	return buf
}

func PrependHeader(data []byte, id uint8) []byte {
	dataLength := uint32(len(data))
	buf := make([]byte, dataLength+5)
	copy(buf, IntToBytes(dataLength+1))
	buf[4] = id
	copy(buf[5:], data)
	return buf
}
