package rtm2_sdk

import "encoding/binary"

const byteThreshold = 0x8000

func DecodeInt(buff []byte) (int, int) {
	low := binary.LittleEndian.Uint16(buff[0:2])
	if low < byteThreshold {
		return 2, int(low)
	}
	high := uint32(buff[2])
	return 3, int(uint32(low)&0x7FFF + high<<15)
}

func TotalWithLength(size int) (int, int) {
	lenSize := 4
	length := size + lenSize
	if length >= byteThreshold+2 {
		length -= 1
		lenSize -= 1
	} else {
		length -= 2
		lenSize -= 2
	}
	return length, lenSize
}

func EncodeInt(val int, buff []byte) {
	if val >= byteThreshold {
		binary.LittleEndian.PutUint16(buff[0:2], uint16(val&0x7FFF|0x8000))
		buff[2] = uint8(val >> 15)
	} else {
		binary.LittleEndian.PutUint16(buff[0:2], uint16(val&0x7FFF))
	}
}

func EncodeLen(content []byte, buffer []byte) []byte {
	length, lenSize := TotalWithLength(len(content))
	if cap(buffer) < length {
		buffer = make([]byte, length)
	} else {
		buffer = buffer[:length]
	}
	EncodeInt(length, buffer[:lenSize])
	copy(buffer[lenSize:], content)
	return buffer
}
