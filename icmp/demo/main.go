package main

import (
	"fmt"
	"time"
	"github.com/gw123/net_tool/icmp"
	"github.com/gw123/gworker/jobs"
	"github.com/gw123/gworker"
	"github.com/gw123/net_tool/net_utils"
	"flag"
	"strconv"
	"net"
)

type PingJob struct {
	jobs.BaseJob
}

func NewPingJob(ip string) (job *PingJob) {
	job = new(PingJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = jobs.JobFlagNormal
	job.Payload = []byte(ip)
	return
}

func (this *PingJob) DoJob() {
	//fmt.Println("执行任务：", this.WorkerName, string(this.Payload))
	ip := string(this.Payload)
	err := icmp.Ping(ip, 2)
	if err != nil {
		//fmt.Println(err)
		return
	}
	fmt.Println(ip)
}

type ScanJob struct {
	jobs.BaseJob
	Port uint16
}

func NewScanJob(ip string, port uint16) (job *ScanJob) {
	job = new(ScanJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = jobs.JobFlagNormal
	job.Payload = []byte(ip)
	job.Port = port
	return
}

func (this *ScanJob) DoJob() {
	ip := string(this.Payload)
	addr := ip + ":" + strconv.Itoa(int(this.Port))
	conn, err := net.DialTimeout("tcp", addr, time.Second*2)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Println(string(this.Payload))
}

func main() {
	port := flag.Uint("p", 23, "端口")
	method := flag.String("m", "port", "port|ping|printer|box")
	flag.Parse()
	group := gworker.NewWorkerGroup(90)
	group.Start()
	ips := net_utils.GetIpList()

	for _, ip := range ips {
		switch *method {
		case "port":
			job := NewScanJob(ip, uint16(*port))
			group.DispatchJob(job)
			break
		case "ping":
			job := NewPingJob(ip)
			group.DispatchJob(job)
			break
		case "printer":
			job := NewScanJob(ip, uint16(9100))
			group.DispatchJob(job)
			break
		case "box":
			job := NewScanJob(ip, uint16(*port))
			group.DispatchJob(job)
			break
		}
	}

	group.WaitEmpty()
}
