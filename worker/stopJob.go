package worker

import (
	"io"
	"time"
	"github.com/gw123/net_tool/worker/interfaces"
	"fmt"
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

func (this *Job) GetWorkerName() string {
	return this.WorkerName
}

func (this *Job) SetWorkerName(workername string) {
	this.WorkerName = workername
}

func (this *Job) SetPayload(payload []byte) {
	this.Payload = payload
}

func (this *Job) GetPayload() []byte {
	return this.Payload
}

func (this *Job) GetCreatedTime() int64 {
	return this.CreatedTime
}

func (this *Job) SetJobFlag(flag int64) {
	this.Flag = flag
}

func (this *Job) GetJobFlag()int64  {
	return this.Flag
}

func (this *Job) DoJob() {
	fmt.Println("执行Job" + string(this.Payload))
}

func NewJob(payload []byte) (job *Job) {
	job = new(Job)
	job.CreatedTime = time.Now().Unix()
	job.Payload = payload
	job.Flag = interfaces.JobFlagNormal
	return
}
