package cubecode

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Packet represents a Sauerbraten UDP packet.
type Packet struct {
	buf      []byte
	posInBuf int
}

// NewPacket returns a Packet using buf as the underlying buffer.
func NewPacket(buf []byte) *Packet {
	return &Packet{
		buf:      buf,
		posInBuf: 0,
	}
}

// Len returns the length of the packet in bytes.
func (p *Packet) Len() int {
	return len(p.buf)
}

// HasRemaining returns true if there are bytes remaining to be read in the packet.
func (p *Packet) HasRemaining() bool {
	return p.posInBuf < len(p.buf)
}

// SubPacketFromRemaining returns a packet from the bytes remaining in p's buffer.
func (p *Packet) SubPacketFromRemaining() (*Packet, error) {
	return p.SubPacket(p.posInBuf, len(p.buf))
}

// SubPacket returns a part of the packet as new packet. It does not copy the underlying slice!
func (p *Packet) SubPacket(start, end int) (q *Packet, err error) {
	if start < 0 || start > len(p.buf)-1 {
		err = errors.New("cubecode: invalid start index for packet of length " + strconv.Itoa(len(p.buf)))
		return
	}

	if end < 1 || end > len(p.buf) {
		err = errors.New("cubecode: invalid end index for packet of length " + strconv.Itoa(len(p.buf)))
		return
	}

	q = NewPacket(p.buf[start:end])
	return
}

// ReadByte returns the next byte in the packet.
func (p *Packet) ReadByte() (byte, error) {
	if p.posInBuf < len(p.buf) {
		p.posInBuf++
		return p.buf[p.posInBuf-1], nil
	}

	return 0, errors.New("cubecode: buf overread at position " + strconv.Itoa(p.posInBuf))
}

// ReadInt returns the integer value encoded in the next bytes of the packet.
func (p *Packet) ReadInt() (int, error) {
	// n is the size of the buffer
	n := len(p.buf)

	if n < 1 {
		return -1, errors.New("cubecode: buf too short")
	}

	var (
		value int
		err   error
		b     byte
	)

	b, err = p.ReadByte()
	if err != nil {
		return -1, err
	}

	switch b {
	default:
		// most often, th value is only one byte
		// convert to int and keep the sign of the 8-bit representation
		value = int(int8(b))

	case 0x80:
		// value is contained in the next two bytes
		value, err = p.readInt(2)
		// make sure to keep the sign of the 16-bit representation
		value = int(int16(value))

	case 0x81:
		// value is contained in the next four bytes
		value, err = p.readInt(4)

	}

	return value, err
}

// readInt reads n bytes from the packet and decodes them into an int value.
func (p *Packet) readInt(n int) (int, error) {
	var (
		value int
		err   error
		b     byte
	)

	for i := 0; i < n; i++ {
		b, err = p.ReadByte()
		if err != nil {
			return -1, err
		}

		value += int(b) << (8 * uint(i))
	}

	return value, nil
}

// ReadString returns a string of the next bytes up to 0x00.
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

// Matches sauer color codes (sauer uses form feed followed by a digit, e.g. \f3 for red)
var sauerStringsSanitizer = regexp.MustCompile("\\f.")

// SanitizeString returns the string, cleared of sauer color codes like \f3 for red.
func SanitizeString(s string) string {
	s = sauerStringsSanitizer.ReplaceAllLiteralString(s, "")
	return strings.TrimSpace(s)
}
