package main

import (
	"net"
	"fmt"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:9100", time.Second*5)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte("hello wolrd"))
	conn.Close()
}
