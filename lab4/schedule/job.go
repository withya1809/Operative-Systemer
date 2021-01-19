package schedule

import (
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
)

// job keeps track of when the job was started,
// its estimated execution time, currently scheduled time slice,
// and the remaining time. The job also specifies the task to be done
// when run, through the doJob function.
type job struct {
	id        int
	start     time.Time
	estimated time.Duration
	// scheduled represents how long a job is scheduled to run next,
	// either a full quantum or the time remaining for the job. Used by RR and SS.
	scheduled time.Duration
	remaining time.Duration
	doJob     func(time.Duration)
	// ideally these should be factored out in
	// a separate job struct for the stride scheduler
	tickets int
	pass    int
	stride  int
}

func (j job) String() string {
	return fmt.Sprintf("id=%d, est=%v, sch=%v, rem=%v", j.id, j.estimated, j.scheduled, j.remaining)
}

var jobComparer = cmp.Comparer(func(x, y job) bool {
	return x.id == y.id && x.estimated == y.estimated && x.scheduled == y.scheduled && x.remaining == y.remaining
})

type result struct {
	job     // struct embedding
	latency time.Duration
}

func newJob(id int, estimated time.Duration) job {
	return job{
		id:        id,
		estimated: estimated,
		scheduled: estimated,
		remaining: estimated,
		doJob:     time.Sleep,
	}
}

func (j *job) run(durationToRun time.Duration) {
	if j.start.IsZero() {
		// first time we run this job; will be used to calculate latency
		j.start = time.Now()
	}
	j.doJob(durationToRun)
}
