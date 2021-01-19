package schedule

import "sort"

type sjfScheduler struct {
	baseScheduler
}

// newSJFScheduler returns a shortest job first scheduler.
// With this scheduler, jobs are executed in the order of shortest job first.
func newSJFScheduler() *sjfScheduler {
	return &sjfScheduler{
		baseScheduler: baseScheduler{
			runQueue:  make(chan job, queueSize),
			completed: make(chan result, queueSize),
			jobRunner: func(job job) {
				job.run(0)
			},
		},
	}

}

// schedule schedules the provided jobs according to SJF order.
// The tasks with the lowest estimate is scheduled to run first.
func (s *sjfScheduler) schedule(inJobs jobs) {

	sort.SliceStable(inJobs, func(p, q int) bool {
		return inJobs[p].estimated < inJobs[q].estimated
	})

	for _, job := range inJobs {
		s.runQueue <- job //Sender job inn i runQueue
	}
	close(s.runQueue)

}
