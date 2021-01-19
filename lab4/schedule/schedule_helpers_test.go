package schedule

import "time"

var j = func(id int, ts time.Duration) job { return newJob(id, ts) }
var k = func(id, tickets int, ts time.Duration) job { return newSJob(id, tickets, ts) }

type testJobs struct {
	name string
	jobs jobs
}
