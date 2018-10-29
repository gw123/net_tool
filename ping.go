package net_tool

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"sort"
	"math/big"
	"sync"
	"regexp"
)

type HostMap map[string]int

var SuccessMap HostMap
var mutex sync.Mutex

type HostItem struct {
	Ip       string
	UsedTime int
}

type HostArr []HostItem

func (h HostArr) Len() int {
	return len(h)
}

func (h HostArr) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h HostArr) Less(i, j int) bool {
	ip1 := InetAtoN(h[i].Ip)
	ip2 := InetAtoN(h[j].Ip)
	return ip1 < ip2 // 按值排序
}

func InetAtoN(ip string) uint32 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return uint32(ret.Int64())
}

func InetNtoA(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func int2bin(v int)  string{
	var tmp string
	mask := 0x1
	for i :=0 ; i<32 ; i++{
		tmp += strconv.Itoa(mask& (v>>uint(i)))
	}
	return tmp
}

func main() {

	var count int
	var timeout int64
	var size int
	var neverstop bool
	SuccessMap = make(HostMap)

	flag.Int64Var(&timeout, "w", 5000, "等待每次回复的超时时间(毫秒)。")
	flag.IntVar(&count, "n", 2, "要发送的回显请求数。")
	flag.IntVar(&size, "l", 32, "要发送缓冲区大小。")
	flag.BoolVar(&neverstop, "t", false, "Ping 指定的主机，直到停止。")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: ", os.Args[0], "host")
		flag.PrintDefaults()
		flag.Usage()
		os.Exit(1)
	}

	exp1 := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\/(\d{1,2})`)
	exresult := exp1.FindAllStringSubmatch(args[0], -1)
	if exresult == nil {
		fmt.Println("host格式错误 请输入正确的host格式[10.0.0.1/24]")
	}

	ipInt32 := InetAtoN(exresult[0][1])
	offset, _ := strconv.Atoi(exresult[0][2])
	if offset > 32 || offset < 0 {
		fmt.Println("网络掩码设置不正确")
	}
	var mask uint32 = 0xFFFFFFFF
	mask = mask << uint32(32-offset)
	hostName := 0xFFFFFFFF -mask
	//
	//fmt.Println(exresult[0][1:])
	targetNet := ipInt32&mask
	fmt.Printf("正在扫描目标网络  %s \n", InetNtoA(targetNet))

	ch := make(chan int)
	argsmap := map[string]interface{}{}

	argsmap["w"] = timeout
	argsmap["n"] = count
	argsmap["l"] = size
	argsmap["t"] = neverstop
	hosts := make([]string, hostName)
	for i := 0; i < int(hostName); i++ {
		 host := InetNtoA(targetNet+ uint32(i))
		 hosts[i] = host
		 //fmt.Printf("%s \t", host)
	}

	next := true
	//xun'hu
	for times := 1; times < 20 && next; times++ {
		fmt.Printf("第%d次扫描 " , times)
		for _, host := range hosts {
			go ping(host, ch, argsmap)
		}

		for i := 0; i < len(hosts); i++ {
			<-ch
		}

		fmt.Println("是否开始下一次扫描? y/n")
		var input string
		fmt.Scanln(&input)

		if input != "y" {
			next = false
		}
	}

	if len(SuccessMap) ==0 {
		println("未找到存在的主机")
		os.Exit(2)
	}

	sortArr := []HostItem{}
	for h1, t1 := range SuccessMap {
		if t1 == 0 {
			continue
		}
		sortArr = append(sortArr, HostItem{Ip: h1, UsedTime: t1})
	}
	sort.Sort(HostArr(sortArr))
	for _, item := range sortArr {
		fmt.Printf("Host:%-12s \t seq:[ %s ] \n", item.Ip, int2bin(item.UsedTime))
	}
	os.Exit(0)
}

func ping(host string, c chan int, args map[string]interface{}) {
	var count int
	var size int
	var timeout int64
	var neverstop bool
	count = args["n"].(int)
	size = args["l"].(int)
	timeout = args["w"].(int64)
	neverstop = args["t"].(bool)
	//cname, _ := net.LookupCNAME(host)
	net.LookupCNAME(host)
	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		//log.Error("DialTimeout:", err)
		os.Exit(1)
	}
	ip := conn.RemoteAddr()

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
	for count > 0 || neverstop {
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
		checkError(err)

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length+1)
		//fmt.Println("len:" ,ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		if err != nil {
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
		count--
	}

	mutex.Lock()
	if lostN != sendN {
		SuccessMap[ip.String()] = SuccessMap[ip.String()] << 1
		SuccessMap[ip.String()] += 1
	}else {
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
