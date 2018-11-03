package main

import (
	"os"
	"fmt"
	"unsafe"
	"io/ioutil"
	"regexp"
	"strings"
	"errors"
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
	inputKeyBorad := &InputKeyBorad{}
	inputKeyBorad.SetDeviceName("HID 13ba:0001")
	err := inputKeyBorad.Open()
	if err != nil {
		fmt.Println("inputKeyBorad.Open: ", err)
		return
	}
	defer inputKeyBorad.Close()
	for ; ; {
		word, err := inputKeyBorad.Read()
		if err != nil {
			fmt.Println("inputKeyBorad.Read: ", err)
			break
		}
		fmt.Println("input ", word)
	}
}

type InputInterface interface {
	Close() error
	Read() (n byte, err error)
	SetDeviceName(deviceName string)
	GetDeviceName() (string)
	Open()
}

type InputKeyBorad struct {
	DeviceName string
	Filename   string
	fileHandel *os.File
	InputChan  chan byte
}

func (this *InputKeyBorad) GetDeviceName() string {
	return this.DeviceName
}

func (this *InputKeyBorad) SetDeviceName(deviceName string) {
	this.DeviceName = deviceName
}

func (this *InputKeyBorad) Open() error {
	filename, err := this.GetDevicePath()
	if err != nil {
		return err
	}
	this.fileHandel, err = os.OpenFile(filename, os.O_RDONLY, 0660)
	if err != nil {
		return err
	}
	return nil
}

func (this *InputKeyBorad) Close() error {
	err := this.fileHandel.Close()
	return err
}

func (this *InputKeyBorad) Read() (n byte, err error) {
	buffer := make([]byte, 24)
	for ; ; {
		_, err := this.fileHandel.Read(buffer)
		if err != nil {
			return n, err
		}
		inputEvent := *(**InputEvent)(unsafe.Pointer(&buffer))
		//fmt.Println("buffer:", inputEvent.Type, inputEvent.Code, inputEvent.Value)
		switch inputEvent.Type {
		case EV_KEY:
			if inputEvent.Code == KEY_NUMLOCK {
				break
			}
			if inputEvent.Value != 1 {
				return ParseCode(inputEvent.Code), nil
			}
			break
		case EV_SYN:
			//fmt.Println("同步")
			break
		}
	}
	return n, err
}

func (this *InputKeyBorad) GetDevicePath() (string, error) {
	devices, err := os.OpenFile("/proc/bus/input/devices", os.O_RDONLY, 0660)
	if err != nil {
		return "", err
	}

	devicesContent, err := ioutil.ReadAll(devices)
	if err != nil {
		return "", err
	}
	rule := `I: Bus=.*
N: Name="{deviceName}"
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
	rule = strings.Replace(rule, "{deviceName}", this.DeviceName, -1)
	r1, _ := regexp.Compile(rule)
	temps := r1.FindStringSubmatch(string(devicesContent))
	if len(temps) != 2 {
		return "", errors.New("请插入设备")
	}
	filename := "/dev/input/event" + temps[1]
	return filename, nil
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
