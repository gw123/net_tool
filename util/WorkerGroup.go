package util

import (
	"sync"
	"fmt"
)

/***
 * WorkerGroup
 */
type WorkerGroup struct {
	Workers         []*Worker
	Length          int
	waitGroup       *sync.WaitGroup
	WaitingChan     chan *Worker
	WorkerSyncChans []chan int
}

func NewWorkerGroup(size int) (*WorkerGroup) {
	this := new(WorkerGroup)
	this.Workers = make([]*Worker, size)
	this.Length = size
	this.waitGroup = &sync.WaitGroup{}

	this.WaitingChan = make(chan *Worker, size)
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

func (this *WorkerGroup) DispatchJob(job *Job) {
	for {
		worker := <-this.WaitingChan
		worker.AppendJob(job)
		break
	}
}

func (this *WorkerGroup) Stop() {
	for _, worker := range this.Workers {
		job := NewJob([]byte(""))
		job.Flag = JobFlagEnd
		worker.AppendJob(job)
		worker.Stop()
	}
}

func (this *WorkerGroup) Control() {
	for _, worker := range this.Workers {
		go func(worker *Worker) {
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
