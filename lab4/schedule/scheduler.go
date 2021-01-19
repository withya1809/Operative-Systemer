package schedule

import (
	"time"
)

type baseScheduler struct {
	runQueue  chan job
	completed chan result
	jobRunner func(*job)
}

// jobs is a slice of jobs ordered according to some scheduling policies.
type jobs []job

// run starts executing the scheduled jobs from the run queue.
func (s *baseScheduler) run() {

	for job := range s.runQueue { //Looper gjennom runQueue
		s.jobRunner(job)       //Executer jobben
		s.completed <- result{ //Legger til resultatet for jobben inn i completed channel
			job:     job,
			latency: time.Since(job.start),
		}
	}
	close(s.completed)
}

// results returns the channel of results.
// This is primarily used for testing.
func (s *baseScheduler) results() chan result {
	return s.completed
}
