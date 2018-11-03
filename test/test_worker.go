package test

import (
"fmt"
"sync"
"os"
"os/signal"
"syscall"
"github.com/gw123/net_tool/util"
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


	waitGroup := sync.WaitGroup{}
	worker1 := util.NewWorker(&waitGroup, "worker1", false)
	worker1.Begin()
	worker2 := util.NewWorker(&waitGroup, "worker2", false)
	worker2.Begin()

	for i := 1; i <= 10; i++ {
		job := util.NewJob([]byte(fmt.Sprintf("job1 %d", i)))
		worker1.AppendJob(job)
		i++
		job = util.NewJob([]byte(fmt.Sprintf("job2 %d", i)))
		worker2.AppendJob(job)
	}
	waitGroup.Wait()
}

