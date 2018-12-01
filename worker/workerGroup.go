package worker

import (
	"sync"
	"fmt"
	"github.com/gw123/net_tool/worker/interfaces"
)

/***
 * WorkerGroup
 */
type WorkerGroup struct {
	Workers         []*WorkerPipeline
	Length          int
	waitGroup       *sync.WaitGroup
	WaitingChan     chan *WorkerPipeline
	WorkerSyncChans []chan int
}

func NewWorkerGroup(size int) (*WorkerGroup) {
	this := new(WorkerGroup)
	this.Workers = make([]*WorkerPipeline, size)
	this.Length = size
	this.waitGroup = &sync.WaitGroup{}

	this.WaitingChan = make(chan *WorkerPipeline, size)
	this.WorkerSyncChans = make([]chan int, size)
	for index, _ := range this.Workers {
		this.Workers[index] = NewWorker(this.waitGroup, fmt.Sprintf("Worker_%d", index), true)
		//worker.SetWaitGroup(this.waitGroup)
		//worker.SetWorkerName(fmt.Sprintf("worker_%d", index))
		//worker.SetLoopWait(true)
		this.WaitingChan <- this.Workers[index]
		this.WorkerSyncChans[index] = this.Workers[index].runOverChan
	}
	return this
}

func (this *WorkerGroup) DispatchJob(job interfaces.Job) {
	for {
		worker := <-this.WaitingChan
		worker.AppendJob(job)
		break
	}
}

func (this *WorkerGroup) Stop() {
	for _, worker := range this.Workers {
		job := NewJob([]byte(""))
		job.Flag = interfaces.JobFlagEnd
		worker.AppendJob(job)
		worker.Stop()
	}
}

func (this *WorkerGroup) Control() {
	for _, worker := range this.Workers {
		go func(worker *WorkerPipeline) {
			for ; ; {
				<-worker.runOverChan
				this.WaitingChan <- worker
				//fmt.Println("flag", flag)
			}
		}(worker)
	}
}

func (this *WorkerGroup) Start() {
	go this.Control()
	for _, worker := range this.Workers {
		worker.Begin()
	}
}

func (this *WorkerGroup) Wait() {
	this.waitGroup.Wait()
}
