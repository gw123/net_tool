package main

import (
	"fmt"
	//"time"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
	"io"
)

type Job struct {
	WorkerName  string
	CreatedTime int64
	UpdatedTime int64
	Flag        int64
	JobType     string
	Payload     []byte
	Response    []byte
	Input       io.WriteCloser
	Output      io.ReadCloser
}

func (this *Job) SetWriteCloser(input io.WriteCloser) {
	this.Input = input
}

func (this *Job) SetReadCloser(ouput io.ReadCloser) {
	this.Output = ouput
}

func NewJob(payload []byte) (job *Job) {
	job = new(Job)
	job.CreatedTime = time.Now().Unix()
	job.Payload = payload
	return
}

const MaxJobs = 1000

type Worker struct {
	WorkerName string
	Jobs       chan *Job

	//读到channel写入数据 woker 暂停执行新的job
	PauseChan chan int
	//读到个channel写入数据 worker 开始执行job
	StartChan chan int
	//向这个channel写入数据 通知worker调用者woker中任务执行完毕
	runOverChan chan int
	//读到这个channel中的数据 说明调用者想要停止执行 woker在处理完当前任务后退出执行 并且要发送runover信号
	StopChan      chan int
	StopFlag      bool
	runFlag       bool
	firstRunFlag  bool
	IsBusy  	  bool
	timeoutChan   <-chan time.Time
	SyncWaitGroup *sync.WaitGroup
	isLoopWait    bool
}

func NewWorker(waitGroup *sync.WaitGroup, workerName string, isLoopWait bool) (worker *Worker) {
	worker = new(Worker)
	worker.Jobs = make(chan *Job, MaxJobs)

	worker.StartChan = make(chan int)
	worker.PauseChan = make(chan int)
	worker.StopChan = make(chan int)
	worker.runOverChan = make(chan int)
	worker.SyncWaitGroup = waitGroup
	worker.SyncWaitGroup.Add(1)
	worker.StopFlag = false
	worker.runFlag = false
	worker.firstRunFlag = false
	worker.isLoopWait = isLoopWait
	worker.WorkerName = workerName
	return
}

func (this *Worker) SetTimeOut(timeout time.Duration) {
	this.timeoutChan = time.After(timeout)
}

func (this *Worker) Stop() {
	this.StopChan <- 1
}

func (this *Worker) Start() {
	this.StartChan <- 1
}

func (this *Worker) Pause() {
	this.PauseChan <- 1
}

func (this *Worker) AppendJob(job *Job) {
	job.WorkerName = this.WorkerName
	this.Jobs <- job
}

func (this *Worker) Begin() {
	if this.firstRunFlag {
		return
	}
	this.firstRunFlag = true
	this.runFlag = true
	go this.control()
	go this.run()
}

func (this *Worker) control() {
	fmt.Println("Control")
	for ; ; {
		if this.StopFlag {
			break
		}
		select {
		case <-this.StartChan:
			fmt.Println("start")
		case <-this.StopChan:
			fmt.Println("stop")
			this.StopFlag = true
			this.runFlag = false
		case <-this.PauseChan:
			fmt.Println("pause")
			this.runFlag = false
		case <-this.timeoutChan:
			fmt.Println("timeoutChan")
			this.StopFlag = true
			this.runFlag = false
		case <-this.runOverChan:
			fmt.Println("runOverChan")
			this.StopFlag = true
			this.runFlag = false
		}
	}
}

func (this *Worker) run() {
	defer func() {
		this.runOverChan <- 1
		this.SyncWaitGroup.Done()
		fmt.Println(this.WorkerName + "任务执行完成3")
	}()

	for ; ; {
		if len(this.Jobs) == 0 && !this.isLoopWait {
			fmt.Println(this.WorkerName + "任务执行完成1")
			break
		}
		if this.StopFlag {
			break
		}

		if !this.runFlag {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		job := <-this.Jobs
		fmt.Println("job over wId: "+job.WorkerName, " "+string(job.Payload))
		fmt.Println(this.WorkerName, len(this.Jobs), this.isLoopWait)
		time.Sleep(time.Millisecond * 1)
	}
	fmt.Println(this.WorkerName + "任务执行完成2")
}

/***
 * WorkGroup
 */
type WorkGroup struct {
	Workers []Worker
	Length int
}

func NewWorkGroup(size int) (*WorkGroup) {
	this := new(WorkGroup)
	this.Workers = make([]Worker, size)
	this.Length = size
	return this
}

func (this *WorkGroup)DispatchJob(job *Job)  {

}


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
	worker1 := NewWorker(&waitGroup, "worker1", false)
	worker1.Begin()
	worker2 := NewWorker(&waitGroup, "worker2", false)
	worker2.Begin()

	for i := 1; i <= 10; i++ {
		job := NewJob([]byte(fmt.Sprintf("job1 %d", i)))
		worker1.AppendJob(job)
		i++
		job = NewJob([]byte(fmt.Sprintf("job2 %d", i)))
		worker2.AppendJob(job)
	}
	waitGroup.Wait()

}
