package main

import (
	"net"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/gw123/net_tool/netInterfaces"
	"fmt"
)

const (
	// 3秒的计时器
	START = "start"
	END   = "end"
)

var log = logrus.New()
// ipNet 存放 IP地址和子网掩码
var ipNet *net.IPNet
// 本机的mac地址，发以太网包需要用到
var localHaddr net.HardwareAddr
var iface string
// 存放最终的数据，key[string] 存放的是IP地址
var data map[string]Info
// 计时器，在一段时间没有新的数据写入data中，退出程序，反之重置计时器
var timer *time.Ticker
var do chan string

type Info struct {
	// IP地址
	Mac net.HardwareAddr
	// 主机名
	Hostname string
	// 厂商信息
	Manuf string
}


func main() {
	netIpLists,err := netInterfaces.GetAllNetIpList()
	if err != nil{
		fmt.Println(err)
		return
	}

	for _, netIpList := range netIpLists{
		fmt.Println(netIpList.Ifce.Name)
		fmt.Println(netIpList.Iplist)

	}

}
