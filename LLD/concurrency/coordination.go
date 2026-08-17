package concurrency

type TaskScheduler struct {
	queue chan func()
}

// if channel size is too small, work accumulates
// if it is too big, memory wastes, experiment to see what fits
// BufferedChannel acts as a Bounded Queue
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		queue: make(chan func(), 1000),
	}
}

func (ts *TaskScheduler) SubmitTask(task func()) {
	ts.queue <- task // Blocks if channel is full (Backpressure)
}

func (ts *TaskScheduler) WorkerLoop() {
	for task := range ts.queue { // Blocks if queue is empty
		task()
	}
}
