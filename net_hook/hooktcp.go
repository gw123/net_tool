package main

import (
	"net"
	"fmt"
	"flag"
	"bufio"
	"time"
	"github.com/gw123/net_tool/net_utils"
)

func PrintSelfInfo() {
	ips := net_utils.GetLocalIpList(nil)
	fmt.Println("Machine ip:")
	for _, ip := range ips {
		fmt.Println(ip)
	}
	fmt.Println()
}

func main() {
	PrintSelfInfo()

	port := flag.String("port", "9100", "设置监听的端口")
	flag.Parse()
	addr := "0.0.0.0:" + *port
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("server listen on ", addr)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go DoConn(conn)
	}
}

func DoConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	fmt.Printf("New Conn %s\n", conn.RemoteAddr().String())
	addr := conn.RemoteAddr().String()
	var line []byte
	var err error
	for err == nil {
		conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		line, _, err = reader.ReadLine()
		if err != nil {
			fmt.Println("Error ", err.Error())
			return
		}
		fmt.Printf("[%s] %s \n", addr, string(line))
	}
	fmt.Println("Close Conn ", addr)
}
