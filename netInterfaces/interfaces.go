package netInterfaces

import (
	"net"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"strconv"
	"math"
	"github.com/google/gopacket/pcap"
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

type IP uint32

// 根据IP和mask换算内网IP范围
func GetNetIpList(ipNet *net.IPNet) []string {
	ip := ipNet.IP.To4()
	var min, max IP
	var data []string
	for i := 0; i < 4; i++ {
		b := IP(ip[i] & ipNet.Mask[i])
		min += b << ((3 - uint(i)) * 8)
	}
	one, _ := ipNet.Mask.Size()
	max = min | IP(math.Pow(2, float64(32-one))-1)
	// max 是广播地址，忽略
	// i & 0x000000ff  == 0 是尾段为0的IP，根据RFC的规定，忽略
	for i := min; i < max; i++ {
		if i&0x000000ff == 0 {
			continue
		}
		data = append(data, InetNtoA(int64(i)))
	}
	return data
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


func GetAllNetIpList() (netIpLists []*NetIpList, err error) {
	ifces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, ifce := range ifces {
		if ifce.Flags&net.FlagUp == 0 {
			continue
		}
		//过滤虚拟网卡
		if strings.Contains(ifce.Name, "VMware") ||
			strings.Contains(ifce.Name, "vmnet") ||
			strings.Contains(ifce.Name, "vmnet") ||
			strings.Contains(ifce.Name, "VirtualBox") ||
			strings.Contains(ifce.Name, "docker") {
			continue
		}

		addrs, err := ifce.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok &&
				ipnet.IP.To4() != nil &&
				!ipnet.IP.IsLoopback() &&
				!strings.HasPrefix(ipnet.IP.String(), "169.254") {
				iplist := GetNetIpList(ipnet)
				netIpList := NewNetIPList(ifce)
				netIpList.Iplist = append(netIpList.Iplist, iplist...)
				netIpLists = append(netIpLists, netIpList)
			}
		}
	}
	return netIpLists, nil
}


func FindIfIdGetPrefixIp(prefix string) string {
	//  获取网卡列表
	var devices []pcap.Interface
	devices, _ = pcap.FindAllDevs()
	for _, d := range devices {
		for _, addr := range d.Addresses {
			if ip4 := addr.IP.To4(); ip4 != nil {
				ip4.Mask(addr.Netmask)
				if strings.HasPrefix(ip4.String(), prefix) {
					//data, _ := json.MarshalIndent(d, "", "  ")
					//fmt.Println(string(data))
					return d.Name
				}
			}
		}
	}
	return ""
}

func FindPcapIfNameByIp(ip string) string {
	//  获取网卡列表
	var devices []pcap.Interface
	devices, _ = pcap.FindAllDevs()

	for _, d := range devices {
		for _, addr := range d.Addresses {
			if ip4 := addr.IP.To4(); ip4 != nil {
				if net.ParseIP(ip).Mask(addr.Netmask).Equal(ip4.Mask(addr.Netmask).To4()) {
					return d.Name
				}
			}
		}
	}
	return ""
}

func FindIfceByIp(ip string) *net.Interface {
	//  获取网卡列表
	devices, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, d := range devices {
		addrs, err := d.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					if ipnet.IP.Mask(ipnet.Mask).Equal(net.ParseIP(ip).Mask(ipnet.Mask)) {
						return &d
					}
				}
			}
		}
	}
	return nil
}

//func FindIfIdGetMac(mac []byte) string {
//	//  获取网卡列表
//	var devices []pcap.Interface
//	devices, _ = pcap.FindAllDevs()
//	for _, d := range devices {
//		if d.Addresses == mac {
//			return d.Name
//		}
//	}
//	return ""
//}