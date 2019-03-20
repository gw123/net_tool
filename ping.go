package main

import (
		"flag"
	"fmt"
	"net"
	"os"
	"time"
	//"sort"
	"sync"
	"strings"
	"sort"
	"github.com/fpay/escpos-go/printer/connection"
	"github.com/gw123/net_tool/net_utils"
)

type HostMap map[string]int

var SuccessMap HostMap
var mutex sync.Mutex

func main() {
	var count int
	var timeout int64
	var size int
	var neverstop bool
	hostArr := make([]string, 0)

	flag.Int64Var(&timeout, "w", 5000, "等待每次回复的超时时间(毫秒)。")
	flag.IntVar(&count, "n", 1, "要发送的回显请求数。")
	flag.IntVar(&size, "l", 32, "要发送缓冲区大小。")
	flag.BoolVar(&neverstop, "t", false, "Ping 指定的主机，直到停止。")
	flag.Parse()
	//args := flag.Args()
	argsmap := map[string]interface{}{}
	argsmap["w"] = timeout
	argsmap["n"] = count
	argsmap["l"] = size
	argsmap["t"] = neverstop

	hosts := net_utils.GetIpList()

	isNeedRoot := func(host string) bool {
		conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") {
				return true
			}
			return false
		}
		conn.Close()
		return false
	}(hosts[0])

	if isNeedRoot {
		fmt.Println("需要管理员权限...")
		return
	}

	fmt.Println("开始扫描.....")
	wg := &sync.WaitGroup{}
	for _, host := range hosts {
		wg.Add(1)
		go func(ip string, c *sync.WaitGroup) {
			defer wg.Done()
			isOk := ping(ip, argsmap)
			mutex.Lock()
			if isOk {
				hostArr = append(hostArr, ip)
			}
			mutex.Unlock()
		}(host, wg)
	}
	wg.Wait()

	sort.Strings(hostArr)
	for _, ip := range hostArr {
		//fmt.Println(ip)
		conn, err := connection.NewNetConnection(ip + ":9100")
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = conn.Write([]byte("###########\n"))
		if err != nil {
			fmt.Println(ip, err)
			continue
		}
	}
}

func ping(host string, args map[string]interface{}) bool {
	var timeout = 5000
	size := 32
	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		//fmt.Println("DialTimeout ",err)
		return false
	}
	var seq int16 = 1
	id0, id1 := genidentifier(host)
	const ECHO_REQUEST_HEAD_LEN = 8

	shortT := -1
	longT := -1
	sumT := 0
	conn.Close()
	{
		var msg []byte = make([]byte, size+ECHO_REQUEST_HEAD_LEN)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
		msg[6], msg[7] = gensequence(seq) //sequence[0], sequence[1]
		length := size + ECHO_REQUEST_HEAD_LEN
		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))
		if err != nil {
			//fmt.Println("DialTimeout --- ", err)
			return false
		}

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length+1)
		//fmt.Println("len:" ,ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		if err != nil {
			return false
		}
		_ = n
		conn.Close()
		var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

		sumT += endduration

		if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			//fmt.Println("对 " + cname + "[" + ip.String() + "]" + " 的请求超时。")
		} else {
			if shortT == -1 {
				shortT = endduration
			} else if shortT > endduration {
				shortT = endduration
			}
			if longT == -1 {
				longT = endduration
			} else if longT < endduration {
				longT = endduration
			}
			//ttl := int(receive[8])
			//			fmt.Println(ttl)
			//fmt.Println("来自 " + cname + "[" + ip.String() + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endduration) + "ms TTL=" + strconv.Itoa(ttl))
		}
		seq++
	}
	//stat(ip.String(), sendN, lostN, recvN, shortT, longT, sumT)
	return true
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func gensequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func stat(ip string, sendN int, lostN int, recvN int, shortT int, longT int, sumT int) {
	fmt.Println()
	fmt.Println(ip, " 的 Ping 统计信息:")
	fmt.Printf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", sendN, recvN, lostN, int(lostN*100/sendN))
	fmt.Println("往返行程的估计时间(以毫秒为单位):")
	if recvN != 0 {
		fmt.Printf("    最短 = %dms，最长 = %dms，平均 = %dms\n", shortT, longT, sumT/sendN)
	}
}
