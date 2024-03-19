package protocol

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const DefaultPayloadLength = 1
const MaxPayloadLength = 4 * 1024 * 1024

type Frame struct {
	buffer []byte
}

func NewFrame() *Frame {
	return &Frame{
		buffer: make([]byte, DefaultPayloadLength),
	}
}

// ReadFrame read a full frame
func (f *Frame) ReadFrame(conn net.Conn) ([]byte, error) {
	frameHeader := make([]byte, FrameHeadLen)
	if num, err := io.ReadFull(conn, frameHeader); num != FrameHeadLen || err != nil {
		return nil, err
	}
	magic := binary.BigEndian.Uint16(frameHeader[:2])
	if magic != Magic {
		return nil, errors.New("invalid magic")
	}

	length := binary.BigEndian.Uint32(frameHeader[8:12])
	if length > MaxPayloadLength {
		return nil, errors.New("payload too large")
	}

	if length > uint32(len(f.buffer)) {
		f.buffer = make([]byte, length)
	}

	if num, err := io.ReadFull(conn, f.buffer); uint32(num) != length || err != nil {
		return nil, err
	}
	return append(frameHeader, f.buffer[:length]...), nil
}
