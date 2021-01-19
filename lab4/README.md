# Lab 4: Scheduling

| Lab 4: | Scheduling |
| ---------------------    | --------------------- |
| Subject:                 | DAT320 Operating Systems and Systems Programming |
| Deadline:                | **September 25, 2020 23:59** |
| Expected effort:         | 5-8 hours |
| Grading:                 | Pass/fail |
| Submission:              | Individually |

## Table of Contents

1. [Introduction](#introduction)
2. [Scheduling](#scheduling)
3. [Task: Implement Scheduling Algorithms](#task-implement-scheduling-algorithms)

## Introduction

In this lab, you will build a job scheduler able to schedule jobs according to different scheduling policies.

## Scheduling

From the lectures you have learned about scheduling.
In this exercise, you will implement three different job schedulers in Go.

| Policy                    | Description                                                                                                                        |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| First In First Out (FIFO) | Schedules jobs in the order of arrival.                                                                                            |
| Shortest Job First (SJF)  | Schedules jobs based on the estimated execution time; runs the shortest jobs first.                                                |
| Round Robin (RR)          | Schedules jobs in the FIFO order, but only for some given time quantum, giving each job a fair share of the processor.              |
| Stride Scheduling (SS)    | Schedules jobs using the Stride Scheduling algorithm in Chapter 9.6, giving each job an exact proportional share of the processor. |

We provide the FIFO scheduler as a template for implementing the more advanced schedulers.
Each scheduler should be able to schedule a list of jobs according to the different scheduling policies given in the table above.

The Round Robin algorithm will need to keep track of the remaining time for each job, since it might not have completed, when its time slice is exhausted.

Stride Scheduling works by giving jobs a *pass* and *stride* value based on each job's allocated *tickets*.
Similar to Round Robin, jobs are scheduled for some given time quantum.
If multiple pass values are the same, use the job with the lowest *stride* value instead.
Note that this is different from the textbook, which say that then the choice is arbitrary.
However, arbitrary choices are not conducive to unit tests.

### Task: Implement Scheduling Algorithms

The code is organized in several files.
Study and familiarize yourself with the code.

The different schedulers that you implement must satisfy the `scheduler` interface in `scheduler_iface.go`.
To that end, you may wish to use the provided `baseScheduler`, as is done in the `fifoScheduler`, when implementing the other schedulers.
However, the other schedulers may need additional inputs and fields in their respective structs, such as the time slice, or `quantum`, as used by RR and SS.

When a job has been **completed** or its current **time slice has been exhausted**, a `result` object should be sent, indicating that the job has executed.
The `result` object should contain a non-zero latency value.
Latency is the duration between the starting time of the job and when the job is completed.

The `scheduler`'s `schedule()` method is responsible for correctly populating the run queue.
The `run()` method is responsible for executing the scheduled jobs from the run queue.
Furthermore, the `run()` method is also responsible for emitting the result on the `completed` channel, which can be accessed by the tests to check the results.
Keep in mind that the scheduler should be able to handle **at least 500 jobs.**
Note that it is important to close the two channels (`completed` and `runQueue`) when all the data has been transmitted, such that the tests or other parts of the program may stop waiting for more values to arrive over these channels.
The `run()` method of the `job` type is responsible for actually executing the job.

Finally, since the stride scheduler works with tickets, stride, and pass values per `job`, you may wish to add the relevant fields to the `job` struct.
That is, unless you are able to find a reasonable way to compose a stride scheduler job, e.g. `sjob` from `job` and the relevant fields and still remain compatible with the current API.
This is non-trivial, and so is not required.

#### Testing the Various Schedulers

You can do preliminary testing of your schedulers locally before pushing to GitHub and Autograder for testing.
That is, to run the various tests in the `scheduler_test.go` file, first cd into the `schedule` directory:

```console
cd lab4/schedule
```

Then use one of these commands:

```console
go test -v -run TestSchedulers/FIFO
go test -v -run TestSchedulers/SJF
go test -v -run TestSchedulers/RR
go test -v -run TestSchedulers/SS
```

If you want to run all tests in one go, use this command:

```console
go test -v -run TestSchedulers
```

Or simply:

```console
go test -v
```

The `-v` flag is used to print verbose output, that is, print a message for each test that passes or fails.
Without the `-v` flag, the command will only print when a test fails.
