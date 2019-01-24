package jobs

import (
	"io"
	"time"
	"github.com/gw123/net_tool/worker/interfaces"
	"fmt"
	"github.com/fpay/escpos-go/printer/connection"
)

type FindPrinterJob struct {
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

func (this *FindPrinterJob) SetWriteCloser(input io.WriteCloser) {
	this.Input = input
}

func (this *FindPrinterJob) SetReadCloser(ouput io.ReadCloser) {
	this.Output = ouput
}

func (this *FindPrinterJob) GetWorkerName() string {
	return this.WorkerName
}

func (this *FindPrinterJob) SetWorkerName(workername string) {
	this.WorkerName = workername
}

func (this *FindPrinterJob) SetPayload(payload []byte) {
	this.Payload = payload
}

func (this *FindPrinterJob) GetPayload() []byte {
	return this.Payload
}

func (this *FindPrinterJob) GetCreatedTime() int64 {
	return this.CreatedTime
}

func (this *FindPrinterJob) SetJobFlag(flag int64) {
	this.Flag = flag
}

func (this *FindPrinterJob) GetJobFlag() int64 {
	return this.Flag
}

func (this *FindPrinterJob) DoJob() {
	//fmt.Println("执行Job" + string(this.Payload))
	remoteAddr := string(this.Payload)
	conn, err := connection.NewNetConnection(remoteAddr + ":9100")
	if err != nil {
		return
	}
	fmt.Println("发现打印机", remoteAddr, "测试打印......")
	_, err = conn.Write([]byte("###########\n"))
	if err != nil {
		fmt.Println(remoteAddr, "打印失败", err)
		return
	}
	fmt.Println(remoteAddr, "打印成功")
}

func NewFindPrinterJob(addr string) (job *FindPrinterJob) {
	job = new(FindPrinterJob)
	job.CreatedTime = time.Now().Unix()
	job.Payload = []byte(addr)
	job.Flag = interfaces.JobFlagNormal
	return
}
