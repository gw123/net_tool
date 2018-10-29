package main

import (
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"golang.org/x/crypto/ssh"
	"time"
	"fmt"
	"log"
	"net"
	"sync"
)

func connect(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 3 * time.Second,
		//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			//fmt.Println("Hostname:", hostname)
			fmt.Println("Remote:", remote)
			//fmt.Println("Key", key)
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

/***
 * 测试ssh间隔
 */
func tryAuth()  {
	pwds := []string{"123","admin","admin","admin","099","456","admin","1231231","099","456","89123","1231231","099"}
	flag := false
	waitGroup := sync.WaitGroup{}
	for index,pwd := range pwds{

		waitGroup.Add(1)
		go func(pwd string, index int) {
			defer waitGroup.Done()
			fmt.Println("try:" + pwd)
			session, err := connect("root", pwd, "192.168.10.1", 22)
			if err != nil {
				log.Println("connect: ", err)
			}else {
				log.Println("Success")
				flag = true
				defer session.Close()
			}
		}(pwd ,index)
		time.Sleep(time.Millisecond*750)
	}
	waitGroup.Wait()
}

//远程执行ssh
func main() {

	session, err := connect("root", "admin", "192.168.0.1", 22)
	if err != nil {
		log.Fatal("connect: ", err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(fd, oldState)

	// excute command
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		log.Fatal(err)
	}

	//session.Run("top")
	session.Run("ash")
	session.Wait()

}
