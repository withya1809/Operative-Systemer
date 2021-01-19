package schedule

import (
	"time"
)

type rrScheduler struct {
	baseScheduler
	quantum time.Duration
}

// newRRScheduler returns a Round Robin scheduler with the time slice, quantum.
func newRRScheduler(quantum time.Duration) *rrScheduler {
	return &rrScheduler{
		baseScheduler: baseScheduler{
			runQueue:  make(chan job, queueSize),
			completed: make(chan result, queueSize),
			jobRunner: func(job job) {
				job.run(job.scheduled)
			},
		},
		quantum: quantum,
	}

}

// schedule schedules the provided jobs in round robin order.
func (s *rrScheduler) schedule(test jobs) {

	var jobs = make([]job, len(test)) //må oppdatere testfil
	copy(jobs, test)

	var counter = 0 //Teller antall ferdig jobs

	for {
		for i, job := range jobs {

			//Håndterer remaining tid, slik at job som ikke er fullført fortsetter neste runde
			if job.remaining >= s.quantum {
				jobs[i].scheduled = s.quantum
				s.runQueue <- jobs[i]
				jobs[i].remaining -= s.quantum

			} else if job.remaining > 0 {
				jobs[i].scheduled = job.remaining
				jobs[i].remaining = 0
				s.runQueue <- jobs[i]
			}

			// Hvis en job er ferdig, øk counter
			if job.remaining == 0 {
				counter += 1
			}

		}
		if counter == len(jobs) {
			break
		}

		counter = 0 // Hvis counter ikke er lik størrelsen på lengden av lista, så betyr det at alle jobs ikke er ferdig, så vi starter å telle på nytt fra neste iterasjon av hele lista.

	}

	close(s.runQueue)
}
