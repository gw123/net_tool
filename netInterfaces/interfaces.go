package netInterfaces

import (
	"net"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"strconv"
)

func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

/***
    ip 格式[10.0.0.1/24]
 */
func FindIpInStr(input string) string {
	exp1 := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\/(\d{1,2})`)
	exresult := exp1.FindAllStringSubmatch(input, -1)
	if exresult == nil {
		return ""
	}
	return exresult[0][1]
}

// 获取自己机器的IP地址
func GetLocalIpList() (ipList []string, err error) {
	ipList = make([]string, 0)
	netAdapers, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, netAdaper := range netAdapers {
		if netAdaper.Flags&net.FlagUp == 0 {
			//断开连接
			continue
		}
		addrs, err := netAdaper.Addrs()
		if err != nil {
			//fmt.Println(err)
			continue
		}
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if strings.HasPrefix(ipnet.IP.String(), "169.254") {
						continue
					}
					//if strings.Contains(netAdaper.Name, "VMware") {
					//	continue
					//}
					ipList = append(ipList, ipnet.IP.String())
				}
			}
		}
	}
	return ipList, nil
}


/***
    获取本机外其他可用IP地址列表
    findLocal 是否查找本地的虚拟网卡
 */
func GetIpList(findLocal bool) (ipList []string, netAdapers []net.Interface, err error) {
	ipList = make([]string, 0)
	netAdapers, err = net.Interfaces()
	if err != nil {
		return nil, netAdapers, err
	}

	for _, netAdaper := range netAdapers {
		if netAdaper.Flags&net.FlagUp == 0 {
			//断开连接
			continue
		}

		addrs, err := netAdaper.Addrs()
		if err != nil {
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

					if !findLocal {
						if strings.Contains(netAdaper.Name, "VMware") {
							continue
						}

						if strings.Contains(netAdaper.Name, "vmnet") {
							continue
						}

						if strings.Contains(netAdaper.Name, "VirtualBox") {
							continue
						}

						if strings.Contains(netAdaper.Name, "docker") {
							continue
						}
					}

					ipInt := InetAtoN(ipnet.IP.String())
					mastInt := big.NewInt(0)
					mastInt.SetBytes(ipnet.Mask)
					mastInt2 := mastInt.Int64()
					totalIp := 0xffffffff - mastInt2
					//mask 子网掩码
					if totalIp > 0x10000 {
						continue
					}

					var i int64
					for i = 1; i < totalIp; i++ {
						newip := InetNtoA((ipInt & mastInt2) + i)
						if ipnet.IP.String() == newip {
							continue
						}
						ipList = append(ipList, newip)
					}
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

type IpSort struct {
	Ip       string
	UsedTime int
}

type IpSortList []IpSort

func (h IpSortList) Len() int {
	return len(h)
}

func (h IpSortList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h IpSortList) Less(i, j int) bool {
	ip1 := InetAtoN(h[i].Ip)
	ip2 := InetAtoN(h[j].Ip)
	return ip1 < ip2 // 按值排序
}
