package lib

import (
	"io"
	"net"
)

const PacketMaxSize = 0xFFFFFF

type Packet struct {
	raw []byte
}

func (p Packet) Size() int {
	head := p.Head()
	return int(uint32(head[0]) | uint32(head[1])<<8 | uint32(head[2])<<16)
}

func (p Packet) Id() byte {
	return p.raw[3]
}

func (p Packet) Data() []byte {
	return p.raw[4:]
}

func (p Packet) Head() []byte {
	return p.raw[:4]
}

func (p Packet) Raw() []byte {
	return p.raw
}

func ReadPacket(from net.Conn) (Packet, error) {
	head := []byte{0, 0, 0, 0}
	packet := Packet{}
	_, err := io.ReadFull(from, head)
	if err != nil {
		return packet, err
	}
	size := int(uint32(head[0]) | uint32(head[1])<<8 | uint32(head[2])<<16)
	data := make([]byte, size)
	if size >= PacketMaxSize {
		total := make([]byte, 0)
		for {
			part := make([]byte, PacketMaxSize)
			_, err := io.ReadFull(from, part)
			if err != nil {
				return packet, err
			}
			total = append(total, part...)
			if len(total) == size {
				data = total
				break
			}
		}
	} else {
		_, err := io.ReadFull(from, data)
		if err != nil {
			return packet, err
		}
	}
	packet.raw = append(head, data...)
	return packet, nil
}
