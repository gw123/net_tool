package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/gw123/net_tool/worker"
	"strconv"
	"github.com/gw123/net_tool/worker/demo/app"
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

	group := worker.NewWorkerGroup(100)
	group.Start()

	for i := 1; i < 255; i++ {
		ipaddr := "192.168.1." + strconv.Itoa(i)+":80"
		job := app.NewCheckIsOpenWRTJob(ipaddr)
		group.DispatchJob(job)
	}

	group.Stop()
	group.Wait()
	//for i := 1; i <= 100; i++ {
	//	work := <- group.WaitingChan
	//	fmt.Println(work)
	//}
}