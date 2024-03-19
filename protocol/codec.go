package protocol

import (
	"bytes"
	"encoding/binary"
)

const FrameHeadLen = 16
const Magic = 499
const Version = 0

type FrameHeader struct {
	Magic        uint16
	Version      uint8
	MsgType      uint8
	ReqType      uint8
	CompressType uint8
	StreamID     uint16
	Length       uint32
	Reserved     uint32
}

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Encode(data []byte) ([]byte, error) {
	totalLen := len(data) + FrameHeadLen
	buffer := bytes.NewBuffer(make([]byte, 0, totalLen))
	frame := FrameHeader{
		Magic:        Magic,
		Version:      Version,
		MsgType:      0,
		ReqType:      0,
		CompressType: 0,
		StreamID:     0,
		Length:       uint32(len(data)),
		Reserved:     0,
	}

	if err := binary.Write(buffer, binary.BigEndian, frame.Magic); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.Version); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.MsgType); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.ReqType); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.CompressType); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.StreamID); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.Length); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, frame.Reserved); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (c *Codec) Decode(frame []byte) []byte {
	return frame[FrameHeadLen:]
}
