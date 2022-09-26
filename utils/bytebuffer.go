package utils

import "encoding/binary"

type ByteBuffer struct {
	data   []byte
	offset uint
}

func NewByteBuffer() ByteBuffer {
	return ByteBuffer{
		data:   []byte{},
		offset: 0,
	}
}

func NewByteBufferFromBuffer(buffer []byte) ByteBuffer {
	return ByteBuffer{
		data:   buffer,
		offset: 0,
	}
}

func (buffer *ByteBuffer) ReadU8() uint8 {
	data := buffer.data[buffer.offset]
	buffer.offset += 1
	return data
}

func (buffer *ByteBuffer) ReadU16() uint16 {
	data := binary.LittleEndian.Uint16(buffer.data[buffer.offset : buffer.offset+2])
	buffer.offset += 2
	return data
}

func (buffer *ByteBuffer) ReadU32() uint32 {
	data := binary.LittleEndian.Uint32(buffer.data[buffer.offset : buffer.offset+4])
	buffer.offset += 4
	return data
}

func (buffer *ByteBuffer) WriteU8(value uint8) {
	buffer.data = append(buffer.data, value)
	buffer.offset += 1
}

func (buffer *ByteBuffer) WriteU16(value uint16) {
	buffer.data = binary.LittleEndian.AppendUint16(buffer.data, value)
	buffer.offset += 2
}

func (buffer *ByteBuffer) WriteU32(value uint32) {
	buffer.data = binary.LittleEndian.AppendUint32(buffer.data, value)
	buffer.offset += 4
}

func (buffer *ByteBuffer) WriteU64(value uint64) {
	buffer.data = binary.LittleEndian.AppendUint64(buffer.data, value)
	buffer.offset += 8
}

func (buffer *ByteBuffer) WriteBytes(value []byte) {
	buffer.data = append(buffer.data, value...)
	buffer.offset += uint(len(value))
}

func (buffer ByteBuffer) GetData() []byte {
	return buffer.data
}
