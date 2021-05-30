package util

type PipeLine interface {

}

type Job interface {

}

type Dispatcher interface {

}

type PipeWorker interface {

}

func CreatePipeWorker()*PipeWorker {
	return nil
}

type WorkerPool interface {
	IsEmpty()bool
	IsClosed()bool
	AddJob(job *Job)
}

type workerPool struct {
	WorkerPool
	pool []*PipeWorker
	closed bool
}

func (w *workerPool) IsEmpty()bool {
	return len(w.pool)<=0
}

func (w *workerPool) IsClosed() bool {
	return w.closed
}

func (w *workerPool)AddJob(job *Job)  {

}
