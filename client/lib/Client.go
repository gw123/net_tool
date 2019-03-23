package lib

import (
	"net"
	"fmt"
	"time"
	"sync"
	"io"
)

type TcpClient struct {
	Conn           net.Conn
	SendChan       chan []byte
	ClientPingChan chan int
	StopChan       chan int
}

func NewTcpCLient(addr string ,wg *sync.WaitGroup) *TcpClient {
	conn, err := net.Dial("tcp", addr)
	//defer conn.Close()
	if err != nil {
		fmt.Println("net.Dial 连接错误: " ,err)
		return nil
	}
	this := &TcpClient{
		Conn:           conn,
		StopChan:       make(chan int, 1),
		ClientPingChan: make(chan int, 100),
		SendChan:       make(chan []byte, 1024),
	}

	wg.Add(1)
	go this.pingAndDataPacket(wg)
	wg.Add(1)
	go this.receivePackets(wg)

	return this
}


func (this *TcpClient) pingAndDataPacket(group *sync.WaitGroup) {
	for {
		select {
		case data, _ := <-this.SendChan:
			this.Conn.Write(data)
		case <-time.Tick(10 * time.Second):
			this.sendHearPacket()
		case <-this.ClientPingChan:
		case <-this.StopChan:
			goto stop

		}
	}

stop:
	group.Done()
	fmt.Println("关闭发送协程")
}

func (this *TcpClient) sendHearPacket() {
	this.Conn.Write([]byte{})
}

func (this *TcpClient) Send(conntent []byte){
	//this.SendChan <- conntent
	this.Conn.Write(conntent)
}

func (this *TcpClient) receivePackets(group *sync.WaitGroup) {
	defer func() {
		fmt.Println("close tcpClient")
		this.StopChan <- 1
		this.Conn.Close()
		if err := recover(); err != nil {
			fmt.Println("readPacket:", err)
		}
	}()
	buffer := make([]byte, 1024)
	timer := time.NewTicker(10 * time.Millisecond)
	for range timer.C {
		count, err := this.Conn.Read(buffer)
		if err == io.EOF {
			goto OnError
		}
		if err != nil {
			//if opErr, ok := err.(*net.OpError); ok {
			//	fmt.Printf("type conversion, error.op: %s, net: %s", opErr.Op, opErr.Net)
			//}
			switch err := err.(type) {
			case *net.OpError:
				if err.Err.Error() == "use of closed network connection" {
					fmt.Printf("服务端已经关闭连接\n")
					goto OnError
				}
				break

			default:
				fmt.Printf("err: %s\n", err)
			}
		}
		fmt.Printf("接收到新的数据包 长度[%d]:\n%s\n", count, string(buffer))
	}

OnError:
	group.Done()
	fmt.Printf("解码数据包结束 over \n")
}
