package arp

import (
	"regexp"
	"fmt"
	"os/exec"
)

func main() {
	macStr, _ := getRawArp()
	//fmt.Println(macStr)
	r, _ := regexp.Compile(`.*?((\d{1,3}.){3}\d{1,3}) .* ((\w{2}:){5}\w{2}) .* ([\w-]+)`)

	temps := r.FindAllStringSubmatch(macStr, -1)
	for _, temp := range temps {
		//fmt.Printf(" mac %s \t,IP %s \t, interface %s\n", temp[3], temp[1], temp[5])
		mac := GetMacByIp(temp[1])
		fmt.Println(mac)
	}
}

func getRawArp() (string, error) {
	//macStr := `IP address       HW type     Flags       HW address            Mask     Device
	//192.168.233.122  0x1         0x2         00:e0:7b:68:01:d8     *        br-lan
	//192.168.233.112  0x1         0x0         54:36:9b:2a:fd:36     *        br-lan
	//192.168.233.228  0x1         0x0         54:36:9b:33:c8:ac     *        br-lan
	//10.0.1.1         0x1         0x2         44:d9:e7:9e:94:1d     *        apcli0
	//10.0.1.105       0x1         0x2         66:09:80:02:bb:d3     *        apcli0
	//10.0.1.219       0x1         0x2         00:02:6f:b9:79:37     *        apcli0
	//`
	//return macStr, nil
	cmd := exec.Command("arp", "-a")
	buf, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func GetMacByIp(ip string) string {
	macStr, _ := getRawArp()

	r, _ := regexp.Compile(`((\d{1,3}.){3}\d{1,3}) .* ((\w{2}:){5}\w{2}) .* ([\w-]+)`)

	temps := r.FindAllStringSubmatch(macStr, -1)
	for _, temp := range temps {
		if temp[1] == ip {
			return temp[3]
		}
	}

	return ""
}

func GetAllIpMacList(ip string) Mac2ipList {
	macStr, _ := getRawArp()
	r, _ := regexp.Compile(`((\d{1,3}.){3}\d{1,3}) .* ((\w{2}:){5}\w{2}) .* ([\w-]+)`)

	temps := r.FindAllStringSubmatch(macStr, -1)
	list := Mac2ipList{}
	for _, temp := range temps {
		fmt.Printf(" mac %s \t,IP %s \t, interface %s\n", temp[3], temp[1], temp[5])
		list.AppendRaw(temp[1], temp[3], temp[5])
	}
	return list
}

type Mac2ip struct {
	Ip     string
	Mac    string
	Interf string
}

type Mac2ipList []Mac2ip

func (this Mac2ipList) AppendRaw(ip, mac, interf string) {
	mac2Ip := Mac2ip{ip, mac, interf}
	this.Append(mac2Ip)
}

func (this Mac2ipList) Append(ip Mac2ip) {
	this = append(this, ip)
}

func (this Mac2ipList) FindByIP(ip string) *Mac2ip {
	for _, item := range this {
		if item.Ip == ip {
			return &item
		}
	}
	return nil
}

func (this Mac2ipList) FindByMac(mac string) *Mac2ip {
	for _, item := range this {
		if item.Mac == mac {
			return &item
		}
	}
	return nil
}

func ArpAskMac(interfa, destIp string) error {
	cmd := exec.Command("arping", "-b", "-c", "2", "-I", interfa, destIp)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
