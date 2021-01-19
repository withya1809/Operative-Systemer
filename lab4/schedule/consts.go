package schedule

import "time"

const (
	// queueSize defines the maximum number of jobs
	// that can be scheduled simultaneously.
	queueSize = 512
)

const (
	ts01  = 1 * time.Millisecond
	ts02  = 2 * time.Millisecond
	ts05  = 5 * time.Millisecond
	ts10  = 10 * time.Millisecond
	ts15  = 15 * time.Millisecond
	ts20  = 20 * time.Millisecond
	ts50  = 50 * time.Millisecond
	ts100 = 100 * time.Millisecond
)

const (
	A = 1
	B = 2
	C = 3
	D = 4
	E = 5
	F = 6
)
