package main

import (
	"fmt"
	//"time"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
)

/****
 * 利用channel同步协程
 */
func main() {
	var falg_stop bool = false
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGUSR2)
		signal.Notify(c, syscall.SIGINT)
		for {
			s := <-c
			//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
			fmt.Println("get signal:", s)
			fmt.Println("完成已处理任务队列后程序结束")
			falg_stop = true
		}
	}()

	ch := make(chan string, 5)
	i := 0
	waitGroup := sync.WaitGroup{}
	allRunOver := make(chan int)

	mutex := sync.Mutex{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for ; ; {
			mutex.Lock()
			if i >= 10000 || falg_stop {
				mutex.Unlock()
				break
			}
			i = i + 1
			ch <- fmt.Sprintf("A:%d", i)
			mutex.Unlock()
			//time.Sleep(time.Millisecond * 100)
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for ; ; {
			mutex.Lock()
			if i >= 10000 || falg_stop {
				mutex.Unlock()
				break
			}
			i = i + 1
			ch <- fmt.Sprintf("B:%d", i)
			mutex.Unlock()
			//time.Sleep(time.Millisecond * 100)
		}
	}()

	go func() {
		waitGroup.Wait()
		//time.Sleep(time.Second)
		for ; len(ch) != 0; {
			time.Sleep(time.Millisecond * 100)
		}
		allRunOver <- 1
	}()

	var data string
	var flag bool = false
	for ; ; {
		if flag {
			break
		}
		fmt.Println("channel len:", len(ch))
		select {
		case data = <-ch:
			fmt.Println(data)
		case <-allRunOver:
			fmt.Println("run over!")
			flag = true
		case <-time.After(time.Second * 10):
			fmt.Println("After 10 second over!")
			flag = true
		}
		//time.Sleep(time.Millisecond * 100)
	}

}
