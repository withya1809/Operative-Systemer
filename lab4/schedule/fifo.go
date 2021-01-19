package schedule

type fifoScheduler struct {
	baseScheduler
}

// newFIFOScheduler returns a FIFO scheduler.
// With this scheduler, jobs are executed in the order of arrival;
// that is, in the order they are provided to the schedule function.
func newFIFOScheduler() *fifoScheduler {
	return &fifoScheduler{
		baseScheduler: baseScheduler{
			runQueue:  make(chan job, queueSize),
			completed: make(chan result, queueSize),
			jobRunner: func(job *job) {
				job.run(0)
			},
		},
	}
}

// schedule schedules the provided jobs according to FIFO order.
func (s *fifoScheduler) schedule(jobs jobs) {
	for _, job := range jobs {
		s.runQueue <- job
	}
	close(s.runQueue)
}
