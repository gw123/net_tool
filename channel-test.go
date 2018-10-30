package main

import (
	"time"
	"io"
	"fmt"
)

type Job struct {
	CreatedTime time.Duration
	UpdatedTime time.Duration
	Flag        int64
	JobType     string
	Payload     []byte
	Input       io.WriteCloser
	Output      io.ReadCloser
}

type Worker struct {
	Jobs chan Job
}

func main() {
	chanInt := make(chan int, 1)
	out := make(chan int, 1)
	go func() {
		go func() {
			time.Sleep(5 * time.Second)
			chanInt <- 5

		}()

		select {
		case x := <-chanInt:
			fmt.Println(x)
		case <-time.After(3 * time.Second):
			fmt.Println("超时了")
		}
		out <- 1
	}()

	<-out
}
