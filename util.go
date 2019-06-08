package net_tool

import (
	"net"
	"bytes"
	"fmt"
)

func ByteToIP(data []byte) net.IP {
	return data
}

func ByteToIp(data []byte) net.IP {
	if len(data) == 4 {
		return net.IP{data[0], data[1], data[2], data[3],}
	}

	if len(data) == 16 {
		return net.IP{
			data[0], data[1], data[2], data[3],
			data[4], data[5], data[6], data[7],
			data[8], data[9], data[10], data[11],
			data[12], data[13], data[14], data[15],}
	}
	return nil
}

func ByteTo16(input []byte) string {
	buf := bytes.Buffer{}
	for _, b := range input {
		flag := byte(b>>4) & 0xf
		if flag == 0 {
			flag = ' '
		} else {
			flag += '0'
		}
		buf.Write([]byte(fmt.Sprintf("%x ", b)))
	}
	return buf.String()
}
