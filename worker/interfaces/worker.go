package interfaces

type Worker interface {
	GetWorkerName() string
	Status()        int
	Isbusy()        bool
	DoJob(job Job)
}
