package main

import (
	"fmt"
	"github.com/gw123/net_tool/netInterfaces"
	"net"
)

func main() {
	ipList, interfaceInfos, err := netInterfaces.GetIpList(true)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, interfaceInfo := range interfaceInfos {
		isPrintFlag := true
		str := fmt.Sprintf("接口:%30s, 硬件地址: %s,\tip:", interfaceInfo.Name, interfaceInfo.HardwareAddr)
		addrs, err := interfaceInfo.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback(){
				if ipnet.IP.IsLoopback(){
					isPrintFlag = false
					continue
				}
				str += ipnet.IP.String() + "\t"
			}
		}
		if isPrintFlag {
			fmt.Println(str)
		}
	}
	fmt.Println("IP List ...")
	for index, ip := range ipList {
		if index%8 == 0 && index != 0 {
			fmt.Println()
		}
		fmt.Printf("%17s", ip)
	}
}
