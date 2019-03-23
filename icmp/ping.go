package icmp

import (
	"time"
	"net"
		"os"
	"github.com/pkg/errors"
	"strings"
)

const ECHO_REQUEST_HEAD_LEN = 8
const ICMP_TYPE_ECHO = 8
const ICMP_TYPE_REPLY = 0
const ECHO_REPLY_HEAD_LEN = 20

const ICMP_ECHO_CODE = 0
const ICMP_REPLY_CODE = 0


func Ping(host string, timeout time.Duration) error {
	//host := "10.0.1.1"
	//var timeout time.Duration = 4
	pid := os.Getpid()
	//fmt.Printf("%x", pid)

	dataSize := 32
	var seq uint8 = 1

	fullLenth := dataSize + ECHO_REQUEST_HEAD_LEN
	frame := make([]byte, fullLenth)

	frame[0] = ICMP_TYPE_ECHO
	frame[1] = ICMP_ECHO_CODE
	frame[2] = 0
	frame[3] = 0
	frame[4], frame[5] = byte(pid>>8&0xff), byte(pid&0xff)
	frame[6], frame[7] = byte(seq>>8&0xff), byte(seq&0xff)
	check := checkSum(frame[0:fullLenth])
	frame[2], frame[3] = byte(check>>8&0xff), byte(check&0xff)

	conn, err := net.DialTimeout("ip:icmp", host, time.Duration(timeout*time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()

	starttime := time.Now()
	conn.SetDeadline(starttime.Add(time.Duration(timeout * time.Second)))
	conn.Write(frame[0:fullLenth])

	var receive []byte = make([]byte, fullLenth+ECHO_REPLY_HEAD_LEN+1)
	n, err := conn.Read(receive)
	if err != nil {
		return err
	}
	_ = n

	if frame[ECHO_REPLY_HEAD_LEN] != ICMP_TYPE_REPLY {
		return errors.New("ICMP REPLY TYPE ERR!")
	}

	if strings.Compare(string(frame[4:fullLenth]), string(receive[ECHO_REPLY_HEAD_LEN+4:fullLenth+ECHO_REPLY_HEAD_LEN])) != 0 {
		return errors.New("ICMP REPLY CONTENT NOT MATCH!")
	}

	return nil
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
