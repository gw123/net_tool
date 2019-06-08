package net_tool

import (
    "time"
    "encoding/binary"
    "math/rand"
    "bytes"
)


// 根据ip生成含mdns请求包，包存储在 buffer里
func Nbns(buffer *Buffer) {
    rand.Seed(time.Now().UnixNano())
    tid := rand.Intn(0x7fff)
    b := buffer.PrependBytes(12)
    binary.BigEndian.PutUint16(b, uint16(tid)) // 0x0000 标识
    binary.BigEndian.PutUint16(b[2:], uint16(0x0010)) // 标识
    binary.BigEndian.PutUint16(b[4:], uint16(1)) // 问题数
    binary.BigEndian.PutUint16(b[6:], uint16(0)) // 资源数
    binary.BigEndian.PutUint16(b[8:], uint16(0)) // 授权资源记录数
    binary.BigEndian.PutUint16(b[10:], uint16(0)) // 额外资源记录数
    // 查询问题
    b = buffer.PrependBytes(1)
    b[0] = 0x20
    b = buffer.PrependBytes(32)
    copy(b, []byte{0x43, 0x4b})
    for i:=2; i<32; i++ {
        b[i] = 0x41
    }
    
    b = buffer.PrependBytes(1)
    // terminator
    b[0] = 0
    // type 和 classIn
    b = buffer.PrependBytes(4)
    binary.BigEndian.PutUint16(b, uint16(33))
    binary.BigEndian.PutUint16(b[2:], 1)
}


func ParseNBNS(data []byte) string {
    var buf bytes.Buffer
    i := bytes.Index(data, []byte{0x20, 0x43, 0x4b, 0x41, 0x41})
    if i < 0 || len(data) < 32 {
        return ""
    }
    index := i + 1 + 0x20 + 12
    // Data[index-1]是在 number of names 的索引上，如果number of names 为0，退出
    if data[index-1] == 0x00 {
        return ""
    }
    for t:= index; ; t++ {
        // 0x20 和 0x00 是终止符
        if data[t] == 0x20 || data[t] == 0x00 {
            break
        }
        buf.WriteByte(data[t])
    }
    return buf.String()
}
