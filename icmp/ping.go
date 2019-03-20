package main

import (
	"os"
	"fmt"
	"net"
	"time"
)

const ECHO_REQUEST_HEAD_LEN = 8
const ICMP_TYPE_ECHO = 8
const ICMP_TYPE_REPLY = 0

const ICMP_ECHO_CODE = 0
const ICMP_REPLY_CODE = 0

func main() {
	host := "10.0.1.1"
	timeout := 2
	pid := os.Getpid()
	dataSize := 32
	seq := 1
	frame := make([]byte, dataSize+ECHO_REQUEST_HEAD_LEN)

	frame[0] = ICMP_TYPE_ECHO
	frame[1] = ICMP_ECHO_CODE
	frame[2] = 0
	frame[3] = 0
	frame[4], frame[5] = byte(pid>>8&0xff), byte(pid&0xff)
	frame[7], frame[8] = byte(seq>>8&0xff), byte(seq&0xff)
	check := checkSum(frame[0 : dataSize+ECHO_REQUEST_HEAD_LEN])
	frame[2], frame[3] = byte(check>>8&0xff), byte(check&0xff)


	conn, err := net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		fmt.Println("DialTimeout --- ", err)
		return
	}

	starttime := time.Now()
	conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))

	const ECHO_REPLY_HEAD_LEN = 20
	var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+dataSize+1)
	//fmt.Println("len:" ,ECHO_REPLY_HEAD_LEN+length)
	n, err := conn.Read(receive)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = n
	conn.Close()
	fmt.Println(receive)
}

func checkSum(msg []byte) uint16 {
	sum := 0
	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}
