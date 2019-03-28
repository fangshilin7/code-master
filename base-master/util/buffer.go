package util

type Buffer struct {
	data []byte
	size int
}

func (buf *Buffer) Add(data []byte) {
	buf.data = append(buf.data, data...)
	buf.size = len(buf.data)
}

func (buf *Buffer) Clear() {
	buf.data = buf.data[:0]
	buf.size = len(buf.data)
}

func (buf *Buffer) Data() []byte {
	return buf.data
}

func (buf *Buffer) Len() int {
	return buf.size
}

func (buf *Buffer) Cap() int {
	return cap(buf.data)
}
