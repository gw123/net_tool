package utils

import (
	"net"
	"fmt"
	"log"
	"strings"
	"math/big"
	"strconv"
)

//func main() {
//	ipList := getIpList()
//	for _, ip := range ipList {
//		fmt.Print(ip, "\t")
//	}
//}
func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func demo2() {
	//exp1 := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\/(\d{1,2})`)
	//exresult := exp1.FindAllStringSubmatch(args[0], -1)
	//if exresult == nil {
	//	fmt.Println("host格式错误 请输入正确的host格式[10.0.0.1/24]")
	//}
	//
	//ipInt32 := uint32(utils.InetAtoN(exresult[0][1]))
	//offset, _ := strconv.Atoi(exresult[0][2])
	//if offset > 32 || offset < 0 {
	//	fmt.Println("网络掩码设置不正确")
	//}

}

func GetIpList(ignoreNetworks []string) (ipList []string) {
	ipList = make([]string, 0)
	netAdapers, err := net.Interfaces()
	if err != nil {
		log.Fatal("无法获取本地网络信息:", err)
	}

	for _, netAdaper := range netAdapers {
		if netAdaper.Flags&net.FlagUp == 0 {
			fmt.Println(netAdaper.Name, "断开连接")
			continue
		}
		//fmt.Println(index, netAdaper)
		addrs, err := netAdaper.Addrs()
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, address := range addrs {
			//fmt.Println(i, address)
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if strings.HasPrefix(ipnet.IP.String(), "169.254") {
						continue
					}

					if strings.Contains(netAdaper.Name, "VMware") {
						continue
					}

					fmt.Println(netAdaper.Name)
					fmt.Print("\tIp2: ", ipnet.IP.String())
					fmt.Println("\tmask: ", ipnet.Mask)
					ipInt := InetAtoN(ipnet.IP.String())
					mastInt := big.NewInt(0)
					mastInt.SetBytes(ipnet.Mask)
					mastInt2 := mastInt.Int64()
					totalIp := 0xffffffff - mastInt2
					startIP := InetNtoA(ipInt & mastInt2)

					fmt.Println("\t网络地址", startIP, "totalIp:", totalIp)
					fmt.Println("\tMac地址:", netAdaper.HardwareAddr)
					var i int64
					for i = 1; i < totalIp; i++ {
						newip := InetNtoA((ipInt & mastInt2) + i)
						//fmt.Print(newip, "\t")
						ipList = append(ipList, newip)
					}
					//fmt.Println()
				}
			}
		}
	}

	return
}

func int2bin(v int) string {
	var tmp string
	mask := 0x1
	for i := 0; i < 32; i++ {
		tmp += strconv.Itoa(mask & (v >> uint(i)))
	}
	return tmp
}

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
