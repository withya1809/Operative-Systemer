package schedule

type scheduler interface {
	schedule(jobs)
	run()
	results() chan result
}
