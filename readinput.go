package main

import (
	"os"
	"fmt"
	"unsafe"
	"io/ioutil"
	"regexp"
)

type InputEvent struct {
	data1 [8]byte
	Type  uint16
	Code  uint16
	Value uint32
}

type InputEventX86 struct {
	data1 [16]byte
	Type  uint16
	Code  uint16
	Value uint32
}

const EV_SYN = 0x00
const EV_KEY = 0x01
const EV_REL = 0x02
const EV_ABS = 0x03
const EV_MSC = 0x04
const EV_SW = 0x05

const KEY_BACKSPACE = 14
const KEY_KPASTERISK = 55
const KEY_NUMLOCK = 69
const KEY_SCROLLLOCK = 70
const KEY_KP7 = 71
const KEY_KP8 = 72
const KEY_KP9 = 73
const KEY_KPMINUS = 74
const KEY_KP4 = 75
const KEY_KP5 = 76
const KEY_KP6 = 77
const KEY_KPPLUS = 78
const KEY_KP1 = 79
const KEY_KP2 = 80
const KEY_KP3 = 81
const KEY_KP0 = 82
const KEY_KPDOT = 83
const KEY_KPENTER = 96
const KEY_KPSLASH = 98

func main() {
	devices, err := os.OpenFile("/proc/bus/input/devices", os.O_RDONLY, 0660)
	if err != nil {
		fmt.Println(err)
		return
	}

	devicesContent, err := ioutil.ReadAll(devices)
	if err != nil {
		fmt.Println(err)
		return
	}

	rule := `I: Bus=.*
N: Name="HID 13ba:0001"
P: Phys=.*
S: Sysfs=.*
U: Uniq=.*
H: Handlers=.*event(\d+).*
B: PROP=.*
B: EV=.*
B: KEY=.*
B: MSC=.*
B: LED=.*
`
	r1, _ := regexp.Compile(rule)
	temps := r1.FindStringSubmatch(string(devicesContent))

	if len(temps) != 2 {
		fmt.Println("请插入设备")
		return
	}
	filename := "/dev/input/event" + temps[1]
	file, err := os.OpenFile(filename, os.O_RDONLY, 0660)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Open " + filename)
	}
	buffer := make([]byte, 24)

	for ; ; {
		_, err := file.Read(buffer)
		inputEvent := *(**InputEvent)(unsafe.Pointer(&buffer))

		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println("buffer:", inputEvent.Type, inputEvent.Code, inputEvent.Value)
		switch inputEvent.Type {
		case EV_KEY:
			if inputEvent.Code == KEY_NUMLOCK {
				break
			}
			if inputEvent.Value != 1 {
				fmt.Printf("Input:%c \n", ParseCode(inputEvent.Code));
			}
			break
		case EV_SYN:
			//fmt.Println("同步")
			break
		}
	}

}

func ParseCode(code uint16) byte {
	switch code {
	case KEY_KP0:
		return '0'
	case KEY_KP1:
		return '1'
	case KEY_KP2:
		return '2'
	case KEY_KP3:
		return '3'
	case KEY_KP4:
		return '4'
	case KEY_KP5:
		return '5'
	case KEY_KP6:
		return '6'
	case KEY_KP7:
		return '7'
	case KEY_KP8:
		return '8'
	case KEY_KP9:
		return '9'
	case KEY_KPDOT:
		return '.'
	case KEY_KPMINUS:
		return '-'
	case KEY_KPPLUS:
		return '+'
	case KEY_BACKSPACE:
		return 'd'
	case KEY_KPENTER:
		return 'y'
	case KEY_KPSLASH:
		return '/'
	case KEY_KPASTERISK:
		return '*'
	}
	return 'e'
}
