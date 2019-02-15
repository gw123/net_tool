package worker

import (
	"fmt"
	//"time"
	"sync"
	"time"
	"github.com/gw123/net_tool/worker/interfaces"
	"github.com/pkg/errors"
)

const MaxJobs = 10

type WorkerPipeline struct {
	WorkerName string
	Jobs       chan interfaces.Job
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

	worker interfaces.Worker
}

func NewWorker(waitGroup *sync.WaitGroup, workerName string, isLoopWait bool) (workerPipeline *WorkerPipeline) {
	workerPipeline = new(WorkerPipeline)
	workerPipeline.Jobs = make(chan interfaces.Job, MaxJobs)
	workerPipeline.StartChan = make(chan int)
	workerPipeline.PauseChan = make(chan int)
	workerPipeline.StopChan = make(chan int)
	workerPipeline.runOverChan = make(chan int)
	workerPipeline.SyncWaitGroup = waitGroup
	workerPipeline.SyncWaitGroup.Add(1)
	workerPipeline.StopFlag = false
	workerPipeline.runFlag = false
	workerPipeline.firstRunFlag = false
	workerPipeline.isBusy = false
	workerPipeline.isLoopWait = isLoopWait
	workerPipeline.WorkerName = workerName
	return
}

func (this *WorkerPipeline) SetWaitGroup(group *sync.WaitGroup) {
	this.SyncWaitGroup = group
	this.SyncWaitGroup.Add(1)
}

func (this *WorkerPipeline) SetWorker(worker interfaces.Worker) {
	this.worker = worker
}

func (this *WorkerPipeline) SetWorkerName(workerName string) {
	this.WorkerName = workerName
}

func (this *WorkerPipeline) SetLoopWait(flag bool) {
	this.isLoopWait = flag
}

func (this *WorkerPipeline) IsBusy() bool {
	if this.isBusy {
		return true
	}
	if len(this.Jobs) == MaxJobs {
		return true
	}
	return false
}

func (this *WorkerPipeline) SetTimeOut(timeout time.Duration) {
	this.timeoutChan = time.After(timeout)
}

func (this *WorkerPipeline) Stop() {
	this.StopChan <- 1
}

func (this *WorkerPipeline) Start() {
	this.StartChan <- 1
}

func (this *WorkerPipeline) Pause() {
	this.PauseChan <- 1
}

func (this *WorkerPipeline) AppendJob(job interfaces.Job) error {
	if this.StopFlag {
		fmt.Println(this.WorkerName, "Run over!!")
		return errors.New("流水线运行结束")
	}
	//job.WorkerName = this.WorkerName
	if len(this.Jobs) == MaxJobs {
		this.isBusy = true
		return errors.New("流水线已满")
	}
	job.SetWorkerName(this.WorkerName)
	this.Jobs <- job
	return nil
}

func (this *WorkerPipeline) Begin() {
	if this.firstRunFlag {
		return
	}
	this.firstRunFlag = true
	this.runFlag = true
	go this.control()
	go this.run()
}

func (this *WorkerPipeline) control() {
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

func (this *WorkerPipeline) run() {
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
			this.isBusy = false
			break
		}
		if this.StopFlag {
			//fmt.Println(this.WorkerName+"任务执行完成step1", "StopFlag")
			this.isBusy = false
			break
		}

		if !this.runFlag {
			time.Sleep(time.Millisecond * 100)
			this.isBusy = false
			continue
		}

		job := <-this.Jobs
		if job.GetJobFlag() == interfaces.JobFlagEnd {
			this.isBusy = false
			break
		}
		job.DoJob()
		//开始执行任务...
		//fmt.Println("Job over wId: "+job.GetWorkerName(), " "+string(job.GetPayload()))
		//fmt.Println(this.WorkerName, len(this.Jobs), this.StopFlag)
		time.Sleep(time.Millisecond * 1)
		this.isBusy = false
		this.runOverChan <- 1
	}

	//fmt.Println(this.WorkerName + "任务执行完成step2")
}
