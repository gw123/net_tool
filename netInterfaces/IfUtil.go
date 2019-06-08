package netInterfaces

import (
	"github.com/google/gopacket"
	"net"
	"github.com/google/gopacket/pcap"
	"time"
	"github.com/google/gopacket/layers"
	"github.com/timest/gomanuf"
	"context"
	"sync"
	"fmt"
	"errors"
	"github.com/gw123/net_tool"
	"strings"
)

type Info struct {
	// IP地址
	Mac net.HardwareAddr
	// 主机名
	Hostname string
	// 厂商信息
	Manuf string
}

//包装发送 arp ,tcp 等请求
type IfUtli struct {
	Ifce          *net.Interface
	LocalIp       net.IP
	localHaddr    net.HardwareAddr
	handle        *pcap.Handle
	packageSource *gopacket.PacketSource
	ctx           context.Context
	stopCtxFun    context.CancelFunc
	HostInfoMap   map[string]Info
}

func NewIfUtli(ifce *net.Interface) *IfUtli {
	this := new(IfUtli)
	this.Ifce = ifce
	this.localHaddr = ifce.HardwareAddr
	this.ctx, this.stopCtxFun = context.WithCancel(context.Background())
	addrs, err := this.Ifce.Addrs()
	this.HostInfoMap = make(map[string]Info, 0)
	if err != nil {
		return this
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				this.LocalIp = ipnet.IP.To4()
				break
			}
		}
	}

	return this
}

// 格式化输出结果
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf
//func PrintData() {
//	var keys IPSlice
//	for k := range data {
//		keys = append(keys, ParseIPString(k))
//	}
//	sort.Sort(keys)
//	for _, k := range keys {
//		d := data[k.String()]
//		mac := ""
//		if d.Mac != nil {
//			mac = d.Mac.String()
//		}
//		fmt.Printf("%-15s %-17s %-30s %-10s\n", k.String(), mac, d.Hostname, d.Manuf)
//	}
//}

// 将抓到的数据集加入到data中，同时重置计时器
func (this *IfUtli) pushData(ip string, mac net.HardwareAddr, hostname, manuf string) {
	// 停止计时器
	var mu sync.RWMutex
	mu.RLock()
	defer func() {
		mu.RUnlock()
	}()
	if _, ok := this.HostInfoMap[ip]; !ok {
		this.HostInfoMap[ip] = Info{Mac: mac, Hostname: hostname, Manuf: manuf}
		return
	}
	info := this.HostInfoMap[ip]
	if len(hostname) > 0 && len(info.Hostname) == 0 {
		info.Hostname = hostname
	}
	if len(manuf) > 0 && len(info.Manuf) == 0 {
		info.Manuf = manuf
	}
	if mac != nil {
		info.Mac = mac
	}
	this.HostInfoMap[ip] = info
}

func (this *IfUtli) GetHostInfo(ip string) *Info {
	var mu sync.RWMutex
	mu.RLock()
	defer func() {
		mu.RUnlock()
	}()
	info, ok := this.HostInfoMap[ip]
	if ok {
		return &info
	}
	return nil
}

//获取本地网卡地址
func (this *IfUtli) GetLocalIp() (ip net.IP, err error) {
	addrs, err := this.Ifce.Addrs()
	if err != nil {
		return ip, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				this.LocalIp = ipnet.IP
				break
			}
		}
	}
	return this.LocalIp, nil
}

//打开接口
func (this *IfUtli) OpenIf() (err error) {
	//fmt.Println("OpenIf ", this.LocalIp.String())
	ifname := FindPcapIfNameByIp(this.LocalIp.String())
	if ifname == "" {
		return errors.New("OpenIf " + this.LocalIp.String() + " can find interface")
	}

	this.handle, err = pcap.OpenLive(ifname, 2048, false, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}

//监听接口上面的消息
func (this *IfUtli) Listen() (err error) {
	if this.handle == nil {
		return errors.New("监听失败 handle is nil ")
	}
	this.handle.SetBPFFilter("arp or (udp and port 5353) or (" + "udp and port 137 and dst host " + this.LocalIp.String() + ")")
	//this.handle.SetBPFFilter("udp and port 5353")
	//this.handle.SetBPFFilter("udp and port 137 and dst host " + this.LocalIp.String())
	ps := gopacket.NewPacketSource(this.handle, this.handle.LinkType())

	go func() {
		for {
			select {
			case <-this.ctx.Done():
				return
			case p := <-ps.Packets():
				this.ParsePacket(p)
			}
		}
	}()
	return nil
}

func (this *IfUtli) Stop() {
	this.stopCtxFun()
	this.handle.Close()
	this.handle = nil
}

//解析数据包
func (this *IfUtli) ParsePacket(packet gopacket.Packet) {
	arp, ok := packet.Layer(layers.LayerTypeARP).(*layers.ARP)
	if ok && arp.Operation == 2 {
		mac := net.HardwareAddr(arp.SourceHwAddress)
		m := manuf.Search(mac.String())
		//pushData(ParseIP(arp.SourceProtAddress).String(), mac, "", m)
		this.pushData(net.IP(arp.SourceProtAddress).String(), mac, "", m)
		ip4 := net.IP(arp.SourceProtAddress).To4()
		if ip4.Equal(this.LocalIp) {
			return
		}

		//解析hostname一次
		hostinfo := this.GetHostInfo(ip4.String())
		fmt.Println("hostinfo : ", ip4.String(), hostinfo)
		if hostinfo == nil || hostinfo.Hostname == "" {
			if strings.Contains(m, "Apple") {
				err := this.SendMdns(ip4, mac)
				if err != nil {
					fmt.Println(err)
				}
				err = this.SendNbns(ip4, mac)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				err := this.SendNbns(ip4, mac)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	_, ok = packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if ok {
		//fmt.Println("source ip : ", ip.SrcIP.String())
		//updPacket, ok := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
		if ok {
			//fmt.Println(updPacket.SrcPort.String(), updPacket.DstPort.String(), net_tool.ByteTo16(updPacket.Payload))
			if len(packet.Layers()) == 4 {
				c := packet.Layers()[3].LayerContents()
				if c[2] == 0x84 && c[3] == 0x00 && c[6] == 0x00 && c[7] == 0x01 {
					// 从网络层(ipv4)拿IP, 不考虑IPv6
					i := packet.Layer(layers.LayerTypeIPv4)
					if i == nil {
						return
					}
					ipv4 := i.(*layers.IPv4)
					ip := ipv4.SrcIP.String()
					h := net_tool.ParseMdns(c)
					//fmt.Println(c)
					if h != "" {
						fmt.Printf("解析MDNS 主机名:%s , IP地址:%s \n", h, ip)
						this.pushData(ip, nil, h, "")
					} else {
						m := net_tool.ParseNBNS(c)
						if len(m) > 0 {
							this.pushData(ip, nil, m, "")
							fmt.Printf("解析NBNS 主机名:%s , IP地址:%s \n", m, ip)
						}
					}
				}
			}
			return
		}
	}
}

//向底层接口发送数据
func (this *IfUtli) WritePacketData(data []byte) error {
	if this.handle == nil {
		return errors.New("WritePacketData handel is nil")
	}
	return this.handle.WritePacketData(data)
}

// 发送arp包
// ip 目标IP地址
func (this *IfUtli) SendArpPackage(ip string) error {
	srcIp := this.LocalIp.To4()
	dstIp := net.ParseIP(ip).To4()
	if dstIp == nil {
		return errors.New("ip 解析出问题")
	}
	fmt.Println("发送Arp请求", dstIp)
	// 以太网首部
	// EthernetType 0x0806  ARP
	ether := &layers.Ethernet{
		SrcMAC:       this.localHaddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}

	a := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     uint8(6),
		ProtAddressSize:   uint8(4),
		Operation:         uint16(1), // 0x0001 arp request 0x0002 arp response
		SourceHwAddress:   this.localHaddr,
		SourceProtAddress: srcIp,
		DstHwAddress:      net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    dstIp,
	}

	buffer := gopacket.NewSerializeBuffer()
	var opt gopacket.SerializeOptions
	gopacket.SerializeLayers(buffer, opt, ether, a)
	outgoingPacket := buffer.Bytes()

	return this.WritePacketData(outgoingPacket)
}

// 发送mdns数据包
func (this *IfUtli) SendMdns(ip net.IP, mhaddr net.HardwareAddr) error {
	srcIp := this.LocalIp.To4()
	dstIp := ip
	ether := &layers.Ethernet{
		SrcMAC:       this.localHaddr,
		DstMAC:       mhaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := &layers.IPv4{
		Version:  uint8(4),
		IHL:      uint8(5),
		TTL:      uint8(255),
		Protocol: layers.IPProtocolUDP,
		SrcIP:    srcIp,
		DstIP:    dstIp,
	}
	bf := net_tool.NewBuffer()
	net_tool.Mdns(bf, ip.String())
	udpPayload := bf.Data
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(60666),
		DstPort: layers.UDPPort(5353),
	}
	udp.SetNetworkLayerForChecksum(ip4)
	udp.Payload = udpPayload // todo
	buffer := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{
		FixLengths:       true, // 自动计算长度
		ComputeChecksums: true, // 自动计算checksum
	}
	err := gopacket.SerializeLayers(buffer, opt, ether, ip4, udp, gopacket.Payload(udpPayload))
	if err != nil {
		return err
	}
	outgoingPacket := buffer.Bytes()
	this.WritePacketData(outgoingPacket)
	return nil
}

func (this *IfUtli) SendNbns(ip net.IP, mhaddr net.HardwareAddr) error {
	fmt.Println("SendNbns packet", ip)
	srcIp := this.LocalIp.To4()
	dstIp := ip.To4()
	ether := &layers.Ethernet{
		SrcMAC:       this.localHaddr,
		DstMAC:       mhaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := &layers.IPv4{
		Version:  uint8(4),
		IHL:      uint8(5),
		TTL:      uint8(255),
		Protocol: layers.IPProtocolUDP,
		SrcIP:    srcIp,
		DstIP:    dstIp,
	}
	bf := net_tool.NewBuffer()
	net_tool.Nbns(bf)
	udpPayload := bf.Data
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(61666),
		DstPort: layers.UDPPort(137),
	}
	udp.SetNetworkLayerForChecksum(ip4)
	udp.Payload = udpPayload
	buffer := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{
		FixLengths:       true, // 自动计算长度
		ComputeChecksums: true, // 自动计算checksum
	}
	err := gopacket.SerializeLayers(buffer, opt, ether, ip4, udp, gopacket.Payload(udpPayload))
	if err != nil {
		return err
	}
	outgoingPacket := buffer.Bytes()

	return this.WritePacketData(outgoingPacket)
}
