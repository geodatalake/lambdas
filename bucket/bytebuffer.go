package bucket

import (
	"errors"
	"io"
)

// Implements ReadSeeker on a byte array
type ByteBuffer struct {
	index int64
	buf   []byte
}

func NewByteBuffer(p []byte) *ByteBuffer {
	return &ByteBuffer{index: 0, buf: p}
}

func (bb *ByteBuffer) max() int {
	return len(bb.buf)
}

func (bb *ByteBuffer) Read(p []byte) (int, error) {
	if int(bb.index) == bb.max() {
		return 0, io.EOF
	}
	reqLen := len(p)
	max := int(bb.index) + reqLen
	if max > bb.max() {
		max = bb.max()
	}
	start := int(bb.index)
	bb.index = int64(max)
	return copy(p, bb.buf[start:max]), nil
}

func (bb *ByteBuffer) Seek(offset int64, whence int) (int64, error) {
	var newIndex int64
	switch whence {
	case io.SeekStart:
		newIndex = offset
	case io.SeekCurrent:
		newIndex = bb.index + offset
	case io.SeekEnd:
		newIndex = int64(bb.max()) - offset
	default:
		return -1, errors.New("Invalid whence")
	}
	if newIndex > int64(len(bb.buf)) || newIndex < 0 {
		return -1, errors.New("Invalid Offset")
	}
	bb.index = newIndex
	return bb.index, nil
}
