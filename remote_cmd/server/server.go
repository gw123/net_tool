package server

import (
	"net"
	"github.com/gw123/remote_cmd/types"
	"fmt"
	"strconv"
	"unsafe"
)

type Server struct {
	Ip             string
	Port           int
	Clients        types.TypeClientMap
	ServerListener net.Listener
}

func (this *Server) HandShake(client net.Conn) {

	buffer := make([]byte, 20)

	client.Read(buffer)
	loginServer := *(**types.LoginServer)(unsafe.Pointer(&buffer))

}

func (this *Server) run() {
	addr := "0.0.0.0:" + strconv.Itoa(this.Port)
	ServerListener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	this.ServerListener = ServerListener

	for ; ; {
		client, err := this.ServerListener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go this.HandShake(client)
	}
}

func (this *Server) Start() {
	go this.run()
}
