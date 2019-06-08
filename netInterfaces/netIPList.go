package netInterfaces

import (
	"net"
)

// 记录接口下面可用的ip列表
type NetIpList struct {
	Ifce   net.Interface
	Iplist []string
}

func NewNetIPList(ifce net.Interface) *NetIpList {
	this := new(NetIpList)
	this.Iplist = make([]string, 0)
	this.Ifce = ifce
	return this
}


