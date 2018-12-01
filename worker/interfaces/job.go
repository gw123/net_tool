package interfaces

import "io"

const JobFlagEnd = 0
const JobFlagNormal = 1

type Job interface {
	GetPayload() []byte
	SetPayload([]byte)
	GetCreatedTime() int64
	SetJobFlag(int64)
	GetJobFlag()int64
	SetWorkerName(string2 string)
	GetWorkerName() string
	SetWriteCloser(input io.WriteCloser)
	SetReadCloser(ouput io.ReadCloser)
	DoJob()
}
