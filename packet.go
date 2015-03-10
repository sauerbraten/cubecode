package cubecode

import (
	"errors"
	"regexp"
)

type Packet struct {
	buf      []byte
	posInBuf int
}

func NewPacket(buf []byte) *Packet {
	return &Packet{
		buf:      buf,
		posInBuf: 0,
	}
}

// Returns the amount of bytes in the packet.
func (p *Packet) Len() int {
	return len(p.buf)
}

// Returns true if there are bytes remaining in the packet.
func (p *Packet) HasRemaining() bool {
	return p.posInBuf < len(p.buf)
}

// Returns a part of the packet as new packet. Does not copy the underlying slice!
func (p *Packet) SubPacket(start, end int) *Packet {
	return NewPacket(p.buf[start:end])
}

// Returns the next byte in the packet.
func (p *Packet) ReadByte() (byte, error) {
	if p.posInBuf < len(p.buf) {
		p.posInBuf++
		return p.buf[p.posInBuf-1], nil
	} else {
		return 0, errors.New("cubecode: buf overread!")
	}
}

// Returns the value encoded in the next bytes of the packet.
func (p *Packet) ReadInt() (value int, err error) {
	// n is the size of the buffer
	n := len(p.buf)

	if n < 1 {
		err = errors.New("cubecode: buf too short!")
		return
	}

	var b1 byte
	b1, err = p.ReadByte()
	if err != nil {
		return
	}

	// 0x80 means: value is contained in the next 2 more bytes
	if b1 == 0x80 {

		var b2, b3 byte

		b2, err = p.ReadByte()
		if err != nil {
			return
		}

		b3, err = p.ReadByte()
		if err != nil {
			return
		}

		value = int(b2) + int(b3)<<8
		return
	}

	// 0x81 means: value is contained in the next 4 more bytes
	if b1 == 0x81 {

		var b2, b3, b4, b5 byte

		b2, err = p.ReadByte()
		if err != nil {
			return
		}

		b3, err = p.ReadByte()
		if err != nil {
			return
		}

		b4, err = p.ReadByte()
		if err != nil {
			return
		}

		b5, err = p.ReadByte()
		if err != nil {
			return
		}

		value = int(b2) + int(b3)<<8 + int(b4)<<16 + int(b5)<<24
		return
	}

	// neither 0x80 nor 0x81: value was already fully contained in the first byte
	if b1 > 0x7F {
		value = int(b1) - int(1<<8)
	} else {
		value = int(b1)
	}

	return
}

// Returns a string of the next bytes up to 0x00.
func (p *Packet) ReadString() (s string, err error) {
	var value int
	value, err = p.ReadInt()
	if err != nil {
		return
	}

	for value != 0x00 {
		codepoint := uint8(value)

		s += string(cubeToUni[codepoint])

		value, err = p.ReadInt()
		if err != nil {
			return
		}
	}

	return
}

// Removes C 0x00 bytes and sauer color codes from strings (e.g. \f3 etc. from server description).
func SanitizeString(s string) string {
	re := regexp.MustCompile("\\f.|\\x00")
	return re.ReplaceAllLiteralString(s, "")
}
