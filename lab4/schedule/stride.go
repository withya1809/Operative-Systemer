package schedule

import (
	"sort"
	"time"
)

// strideScheduler may be defined as an alias for rrScheduler; it has the same fields.
type strideScheduler struct {
	baseScheduler
	quantum time.Duration
}

// newStrideScheduler returns a stride scheduler.
// With this scheduler, jobs are executed similar to round robin,
// but with exact proportions determined by how many tickets each job is assigned.
func newStrideScheduler(quantum time.Duration) *strideScheduler {

	return &strideScheduler{
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

// schedule schedules the provided jobs according to a stride scheduler's order.
// The task with the lowest pass is scheduled to run first.
func (s *strideScheduler) schedule(inJobs jobs) {

	//iterer gjennom alle i lista og sjekk om alle jobbene er ferdig

	var counter = 0
	for {

		for _, job := range inJobs {
			if job.remaining == 0 {
				counter += 1
			}

		}
		if counter == len(inJobs) {
			break
		}

		MinPassJob := minPass(inJobs)

		if MinPassJob.remaining >= s.quantum {
			MinPassJob.pass += MinPassJob.stride
			MinPassJob.scheduled = s.quantum
			s.runQueue <- MinPassJob
			MinPassJob.remaining -= s.quantum

		} else if MinPassJob.remaining > 0 {
			MinPassJob.pass += MinPassJob.stride
			MinPassJob.scheduled = MinPassJob.remaining
			MinPassJob.remaining = 0
			s.runQueue <- MinPassJob
		}

		counter = 0

	}

	close(s.runQueue)

}

func minPass(theJobs jobs) job {

	sort.SliceStable(theJobs, func(p, q int) bool { //Sorterer fra lavest til høyest pass value
		return theJobs[p].pass < theJobs[q].pass
	})

	var myslice []job

	for _, job := range theJobs { //Lager en ny liste med alle job som har samme pass value

		if job.pass == theJobs[0].pass {
			myslice = append(myslice, job)
		}
	}

	lowest_stride := myslice[0].stride //starter med første elementet i den nye listen sin stride value, trenger noe for å sammeligne
	lowest_job := myslice[0]           // prøver å finne job med lavest stride value

	for _, sameJob := range myslice {

		if sameJob.stride <= lowest_stride {
			lowest_stride = sameJob.stride
			lowest_job = sameJob

		}
	}

	return lowest_job

	//lowest := 0
	//return lowest
}

//Først execute job med minst pass value (hvis samme pass, ta minst stride)
//Oppdatere job sin pass value ved å addere med stride value til pass value
//Sortere hele listen på nytt
//Execute job med minst pass value (første element i lista)
//Hvis en job er fullført (remaining time == 0), fjern fra lista ?
