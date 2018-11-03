package util

import (
	"fmt"
	//"time"
	"sync"
	"time"
	"io"
)

const JobFlagEnd = 0
const JobFlagNormal = 1

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
	job.Flag = JobFlagNormal
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
	isBusy        bool
	isBusyMutex   sync.Mutex
	timeoutChan   <-chan time.Time
	SyncWaitGroup *sync.WaitGroup
	isLoopWait    bool
}

func NewWorker(waitGroup *sync.WaitGroup, workerName string, isLoopWait bool) (worker *Worker) {
	worker = new(Worker)
	worker.Jobs = make(chan *Job, 1)

	worker.StartChan = make(chan int)
	worker.PauseChan = make(chan int)
	worker.StopChan = make(chan int)
	worker.runOverChan = make(chan int)
	worker.SyncWaitGroup = waitGroup
	worker.SyncWaitGroup.Add(1)
	worker.StopFlag = false
	worker.runFlag = false
	worker.firstRunFlag = false
	worker.isBusy = false

	worker.isLoopWait = isLoopWait
	worker.WorkerName = workerName
	return
}

func (this *Worker) SetWaitGroup(group *sync.WaitGroup) {
	this.SyncWaitGroup = group
	this.SyncWaitGroup.Add(1)
}

func (this *Worker) SetWorkerName(workerName string) {
	this.WorkerName = workerName
}

func (this *Worker) SetLoopWait(flag bool) {
	this.isLoopWait = flag
}

func (this *Worker) IsBusy() bool {
	return this.isBusy
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
	if this.StopFlag {
		fmt.Println(this.WorkerName, "Run over!!")
		return
	}
	job.WorkerName = this.WorkerName
	this.Jobs <- job
	this.isBusy = true
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
	//fmt.Println("Control")
	for ; ; {
		if this.StopFlag {
			break
		}
		select {
		case <-this.StartChan:
			fmt.Println("start")
		case <-this.StopChan:
			//fmt.Println("stop")
			this.StopFlag = true
			this.runFlag = false
		case <-this.PauseChan:
			fmt.Println("pause")
			this.runFlag = false
		case <-this.timeoutChan:
			fmt.Println("timeoutChan")
			this.StopFlag = true
			this.runFlag = false

		}
	}
}

func (this *Worker) run() {
	defer func() {
		this.StopFlag = true
		this.runFlag = false

		if this.SyncWaitGroup != nil {
			this.SyncWaitGroup.Done()
		} else {
			fmt.Println(this.SyncWaitGroup, "not set")
		}
		//this.SyncWaitGroup.Done()
		//fmt.Println(this.WorkerName + "任务执行完成step3")
	}()

	for ; ; {
		if len(this.Jobs) == 0 && !this.isLoopWait {
			//fmt.Println(this.WorkerName + "任务执行完成step1")
			//this.runOverChan <- 1
			this.isBusy = false
			break
		}
		if this.StopFlag {
			//fmt.Println(this.WorkerName+"任务执行完成step1", "StopFlag")
			this.isBusy = false
			//this.runOverChan <- 1
			break
		}

		if !this.runFlag {
			time.Sleep(time.Millisecond * 100)
			this.isBusy = false

			continue
		}

		job := <-this.Jobs
		if job.Flag == JobFlagEnd {
			//this.runOverChan <- 1
			this.isBusy = false
			break
		}
		fmt.Println("Job over wId: "+job.WorkerName, " "+string(job.Payload))
		//fmt.Println(this.WorkerName, len(this.Jobs), this.StopFlag)
		time.Sleep(time.Millisecond * 1)
		this.isBusy = false
		this.runOverChan <- 1

	}

	//fmt.Println(this.WorkerName + "任务执行完成step2")
}
