package cubecode

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
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
func (p *Packet) SubPacket(start, end int) (q *Packet, err error) {
	if start < 0 || start > len(p.buf)-1 {
		err = errors.New("cubecode: invalid start index for packet of length " + strconv.Itoa(len(p.buf)) + "!")
		return
	}

	if end < 1 || end > len(p.buf) {
		err = errors.New("cubecode: invalid end index for packet of length " + strconv.Itoa(len(p.buf)) + "!")
		return
	}

	q = NewPacket(p.buf[start:end])
	return
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

	var b byte
	b, err = p.ReadByte()
	if err != nil {
		return
	}

	moreBytes := 0

	switch b {
	case 0x80:
		moreBytes = 2
	case 0x81:
		moreBytes = 4
	default:
		// neither 0x80 nor 0x81: value was already fully contained in the first byte
		value = int(b)
		return
	}

	var nextByte byte

	for i := 0; i < moreBytes; i++ {
		nextByte, err = p.ReadByte()
		if err != nil {
			return
		}

		value += int(nextByte) << (8 * uint(i))
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

// Matches sauer color codes (sauer uses form feed followed by a digit, e.g. \f3 for red)
var sauerStringsSanitizer = regexp.MustCompile("\\f.")

// Returns the string, cleared of sauer color codes
func SanitizeString(s string) string {
	s = sauerStringsSanitizer.ReplaceAllLiteralString(s, "")
	return strings.TrimSpace(s)
}
