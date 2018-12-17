package main

import (
	"github.com/gw123/net_tool/utils"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
	//"sort"
	"sync"
)

type HostMap map[string]int
var SuccessMap HostMap
var mutex sync.Mutex

func main() {
	var count int
	var timeout int64
	var size int
	var neverstop bool
	SuccessMap = make(HostMap)

	flag.Int64Var(&timeout, "w", 5000, "等待每次回复的超时时间(毫秒)。")
	flag.IntVar(&count, "n", 1, "要发送的回显请求数。")
	flag.IntVar(&size, "l", 32, "要发送缓冲区大小。")
	flag.BoolVar(&neverstop, "t", false, "Ping 指定的主机，直到停止。")
	flag.Parse()
	//args := flag.Args()
	ch := make(chan int)
	argsmap := map[string]interface{}{}

	argsmap["w"] = timeout
	argsmap["n"] = count
	argsmap["l"] = size
	argsmap["t"] = neverstop

	hosts := utils.GetIpList(nil)

	//xun'hu
	fmt.Println("开始扫描.....")
	for _, host := range hosts {
		go ping(host, ch, argsmap)
	}

	for i := 0; i < len(hosts); i++ {
		<-ch
	}

	if len(SuccessMap) == 0 {
		println("未找到存在的主机")
		os.Exit(2)
	}

	//sortArr := []HostItem{}
	for h1, t1 := range SuccessMap {
		if t1 == 0 {
			continue
		}
		fmt.Println("Ip: ", h1)
		//sortArr = append(sortArr, HostItem{Ip: h1, UsedTime: t1})
	}
	//sort.Sort(HostArr(sortArr))
	//for _, item := range sortArr {
	//	fmt.Printf("Host:%-12s \t seq:[ %s ] \n", item.Ip, int2bin(item.UsedTime))
	//}
	//os.Exit(0)
}

func ping(host string, c chan int, args map[string]interface{}) {
	var timeout = 5000
	size := 32
	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		println("DialTimeout: ", host, err)
		return
	}
	ip := conn.RemoteAddr()
	//cname, _ := net.LookupCNAME(host)
	//fmt.Println(cname)
	//fmt.Println("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")
	var seq int16 = 1
	id0, id1 := genidentifier(host)
	const ECHO_REQUEST_HEAD_LEN = 8

	sendN := 0
	recvN := 0
	lostN := 0
	shortT := -1
	longT := -1
	sumT := 0
	conn.Close()
	{
		sendN++
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
			fmt.Println("DialTimeout ", err)
			return
		}

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length+1)
		//fmt.Println("len:" ,ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		if err != nil {
			//可以记录日志
			//fmt.Println("conn.Read ", err)
			//log.Error(err)
		}
		_ = n
		conn.Close()
		var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

		sumT += endduration

		//time.Sleep(1000 * 1000 * 1000)

		if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			lostN++
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
			recvN++
			//ttl := int(receive[8])
			//			fmt.Println(ttl)
			//fmt.Println("来自 " + cname + "[" + ip.String() + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endduration) + "ms TTL=" + strconv.Itoa(ttl))
		}

		seq++
	}

	mutex.Lock()
	if lostN != sendN {
		SuccessMap[ip.String()] = SuccessMap[ip.String()] << 1
		SuccessMap[ip.String()] += 1
	} else {
		SuccessMap[ip.String()] = SuccessMap[ip.String()] << 1
	}
	//SuccessMap[ip.String()] |= 1<<31
	mutex.Unlock()
	//stat(ip.String(), sendN, lostN, recvN, shortT, longT, sumT)
	c <- 1
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
