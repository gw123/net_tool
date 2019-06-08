package main

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket"
	"github.com/fpay/foundation/charset"
	"strings"
	"encoding/json"
	"os"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/layers"
)

func main() {
	//  获取 libpcap 的版本
	//version := pcap.Version()
	//fmt.Println(version)
	//  获取网卡列表
	var devices []pcap.Interface
	devices, _ = pcap.FindAllDevs()
	//fmt.Println(devices)
	for _, device := range devices {
		//fmt.Println(device.Name)
		//fmt.Println(device.Description)
		fmt.Println(device.Addresses)
	}
	deviceName := findNetName("172.")
	fmt.Println("deviceName" ,deviceName)

	handle, err := pcap.OpenLive(deviceName, 65535, false, -1)
	if err != nil {
		data, _ := charset.GBKToUTF8([]byte(err.Error()))
		fmt.Println(string(data))
		return
	}
	defer handle.Close()

	handle.SetBPFFilter("tcp")
	handle.SetBPFFilter("port 80")
	packetSource := gopacket.NewPacketSource(
		handle,
		handle.LinkType(),
	)

	for packet := range packetSource.Packets() {
		//fdmt.Println(packet)
	}
	dumpFile("abc.pcap" ,packetSource.Packets())
}

func findNetName(prefix string) string {
	//  获取网卡列表
	var devices []pcap.Interface
	devices, _ = pcap.FindAllDevs()
	for _, d := range devices {
		for _, addr := range d.Addresses {
			if ip4 := addr.IP.To4(); ip4 != nil {
				if strings.HasPrefix(ip4.String(), prefix) {
					data, _ := json.MarshalIndent(d, "", "  ")
					fmt.Println(string(data))
					return d.Name
				}
			}
		}
	}
	return ""
}

func dumpFile(filename string,packetChan chan gopacket.Packet)  {
	dumpFile, _ := os.Create(filename)
	defer dumpFile.Close()
	//  准备好写入的 Writer
	packetWriter := pcapgo.NewWriter(dumpFile)
	packetWriter.WriteFileHeader(
		65535,  //  Snapshot length
		layers.LinkTypeEthernet,
	)
	//  写入包
	for packet := range packetChan {
		packetWriter.WritePacket(
			packet.Metadata().CaptureInfo,
			packet.Data(),
		)
	}
}
