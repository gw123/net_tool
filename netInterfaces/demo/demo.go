package main

import (
	"github.com/gw123/net_tool/netInterfaces"
	"fmt"
	"net"
	"github.com/fpay/foundation/charset"
	"flag"
	"github.com/gw123/net_tool/net_log"
)

func tstList() {
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
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
				if ipnet.IP.IsLoopback() {
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

func main() {
	ifname := flag.String("ifname", "", "接口名称")
	dstip := flag.String("dst", "", "目标")
	flag.Parse()

	var ifce *net.Interface
	if *ifname != "" {
		ifce, _ = net.InterfaceByName(*ifname)
	}

	if ifce == nil {
		if *dstip == "" {
			fmt.Println("请输入接口名称或者目标地址")
			return
		}
		ifce = netInterfaces.FindIfceByIp(*dstip)
	}

	if ifce == nil {
		fmt.Println("找不到这样的接口")
		return
	}

	ifUtil := netInterfaces.NewIfUtli(ifce)
	err := ifUtil.OpenIf()
	if err != nil {
		str, err := charset.GBKToUTF8([]byte(err.Error()))
		if err != nil {
			fmt.Println(err)
			return
		}
		net_log.Logout("info", string(str))
		return
	}

	err = ifUtil.Listen()
	if err != nil {
		fmt.Println("Listen : ", err.Error())
		return
	}

	err = ifUtil.SendArpPackage(*dstip)
	if err != nil {
		fmt.Println("SendArpPackage", err)
	} else {
		fmt.Println("发送完毕")
	}

	select {}
}
