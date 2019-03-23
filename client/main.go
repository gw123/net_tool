package main

import (
	"sync"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/gw123/net_tool/clientol/client/lib"
)

func main() {


	wg := new(sync.WaitGroup)
	client := lib.NewTcpCLient("10.0.1.150:9100" , wg)
	if client == nil{
		fmt.Println("连接失败")
		os.Exit(1)
	}
	str ,err := ioutil.ReadFile("z.txt");
	if err != nil{
		fmt.Println(err)
		os.Exit(1);
	}
	//str := "Hello world \n \n\n\n\n\n\n\n\n\n\n\n\n"
	client.Send([]byte(str))
	//fmt.Println(str)
	wg.Wait()
}
