package schedule

import (
	"fmt"
	"testing"
)

var fifoOrder = [][]int{
	{},
	{1, 2},
	{1, 2, 3},
	{1, 2, 3, 4, 5},
}

var rr5Order = [][]int{
	{},
	{1, 2, 1, 2},
	{1, 2, 3, 1, 2, 3},
	{1, 2, 3, 4, 5, 1, 2, 3, 4, 5},
}

var sjfOrder = [][]int{
	{},
	{1, 2},
	{1, 2, 3},
	{1, 2, 3, 4, 5},
}

var theJobs = []testJobs{
	{"No jobs", jobs{}},
	{"Two jobs", jobs{j(1, ts10), j(2, ts10)}},
	{"Three jobs", jobs{j(1, ts10), j(2, ts10), j(3, ts10)}},
	{"Five jobs", jobs{j(1, ts10), j(2, ts10), j(3, ts10), j(4, ts10), j(5, ts10)}},
}

// more jobs is the same as above + some more (used for SJF)
var moreJobs = []testJobs{
	{"No jobs", jobs{}},
	{"Two jobs", jobs{j(1, ts10), j(2, ts10)}},
	{"Three jobs", jobs{j(1, ts10), j(2, ts10), j(3, ts10)}},
	{"Five jobs", jobs{j(1, ts10), j(2, ts10), j(3, ts10), j(4, ts10), j(5, ts10)}},
}

var strideJobs = []testJobs{
	{"No jobs", jobs{}},
	{"ABC jobs", jobs{k(A, 100, ts20), k(B, 50, ts20), k(C, 250, ts20)}},
}

var strideOrder = [][]int{
	{},
	// only four of each job, since time slice is 5 ms and each job is only 20 ms long
	{C, A, B, C, C, A, C, A, B, A, B, B},
}

var schedulerTypes = []struct {
	name            string
	constructorName string
	createScheduler func() scheduler
	jobs            []testJobs
	order           [][]int
}{
	{"FIFO", "newFIFOScheduler", func() scheduler { return newFIFOScheduler() }, theJobs, fifoOrder},
	{"RR(5)", "newRRScheduler", func() scheduler { return newRRScheduler(ts05) }, theJobs, rr5Order},
	{"SJF", "newSJFScheduler", func() scheduler { return newSJFScheduler() }, moreJobs, sjfOrder},
	{"SS(5)", "newStrideScheduler", func() scheduler { return newStrideScheduler(ts05) }, strideJobs, strideOrder},
}

func TestSchedulers(t *testing.T) {
	for _, sch := range schedulerTypes {
		for i, test := range sch.jobs {
			name := fmt.Sprintf("%s/%s", sch.name, test.name)
			t.Run(name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s: panicked; this can happen for several reasons;\n"+
							"Did you implement a '%v' function that returns the corresponding scheduler?\n", name, sch.constructorName)
					}
				}()

				// copy the test.jobs to allow implementations with side-effects
				theJobs := make(jobs, len(test.jobs))
				copy(theJobs, test.jobs)

				sched := sch.createScheduler()
				sched.schedule(theJobs)
				sched.run()

				j := 0
				for res := range sched.results() {
					if j >= len(sch.order[i]) {
						t.Errorf("test schedule '%s' failed; too many results, got %d, want only %d jobs", test.name, j+1, len(sch.order[i]))
						break
					}
					if res.id != sch.order[i][j] {
						t.Errorf("test schedule '%s' failed; got job %d, want job %d", test.name, res.id, sch.order[i][j])
					}
					j++
				}
				if j != len(sch.order[i]) {
					t.Errorf("test schedule '%s' failed; too few results, got %d jobs, want %d jobs", test.name, j, len(sch.order[i]))
				}
			})
		}
	}
}
