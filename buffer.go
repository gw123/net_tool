package net_tool

type Buffer struct {
	Data  []byte
	start int
}

func (b *Buffer) PrependBytes(n int) []byte {
	length := cap(b.Data) + n
	newData := make([]byte, length)
	copy(newData, b.Data)
	b.start = cap(b.Data)
	b.Data = newData
	return b.Data[b.start:]
}

func NewBuffer() *Buffer {
	return &Buffer{

	}
}

// 反转字符串
func Reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}