# Lab 6: Concurrency and Parallelism

| Lab 6: | Concurrency and Parallelism |
| ---------------------    | --------------------- |
| Subject:                 | DAT320 Operating Systems and Systems Programming |
| Deadline:                | **October 25, 2020 23:59** |
| Expected effort:         | 10-15 hours |
| Grading:                 | Pass/fail |
| Submission:              | Group |

## Table of Contents

1. [Introduction](#introduction)
2. [Recommended Reading](#recommended-reading)
3. [Task: Parallel Execution](#task-parallel-execution)
4. [High-level Synchronization & Data Race Detection](#high-level-synchronization-&-data-race-detection)
5. [Task: Multiple Choice Questions About Data Races](#task-multiple-choice-questions-about-data-races)
6. [Task: Implement Thread-Safe Stacks](#task-implement-thread-safe-stacks)
7. [CPU and Memory Profiling](#cpu-and-memory-profiling)
8. [Task: Multiple Choice Questions About CPU and Memory Profiling](#task-multiple-choice-questions-about-cpu-and-memory-profiling)
9. [Task: Benchmarking and Profiling Go Programs](#task-benchmarking-and-profiling-go-programs)

## Introduction

This lab assignment is divided into three parts and deals with two separate programming tools.
In the first part, you will work with parallel execution using goroutines.
The second part will focus on high-level synchronization techniques and will give an introduction to Go’s built-in data race detector.
You will use two different techniques to ensure synchronization to a shared data structure.
The third part of the lab deals with CPU and memory profiling.
We will analyze different implementations of a simple data structure.

### Recommended Reading

An important resource for this assignment is the [`sync` package](https://golang.org/pkg/sync/) of the standard library.

Further, you may need to read again some of the material from the list of resources from Introduction to Go Programming assignment.
In particular you will want to take a look at chapters about concurrency, goroutines, channels, and synchronization primitives, such as mutex locks and wait groups.
Here are some direct pointers:

* [The Go Programming Language (book)](http://www.gopl.io): Chapters 7 and 9.
* [Collection of Videos about Go](https://github.com/golang/go/wiki/GoTalks), specifically this video about [Concurrency](https://youtu.be/f6kdp27TYZs).
* [Golang Tutorial Series](https://golangbot.com/learn-golang-series/): Sections on Concurrency.

## Task: Parallel Execution

Parallel execution is often quite simple in Go.
In this part you will implement a function to perform parallel word count.
We have provided you with a large text file (`mobydick.txt`) and three simple functions in the file `wc.go` you should use:

* The function `loadMoby()` loads the file and returns it as a `[]byte`.
* The function `wordCount()` counts the words in a `[]byte`.
* And the function `shardSlice()` splits the provided `[]byte` into sub-slices that can be counted separately.

1. Your first task is to implement a function called `parallelWordCount()` with the following signature:

    ```go
    func parallelWordCount(input []byte) (words int)
    ```

    This function *must count* the words in `mobydick.txt` file using multiple goroutines; typically as many as there are CPU cores on your machine.
    You can reuse the provided functions as you see fit.
    The `TestParallelWordCount` must be passed, meaning that it should return the same number of words as the sequential version of word count (the `wordCount()` function).

2. Perform benchmark tests using the provided benchmark tests. Run as follows:

   ```console
   go test -v -run=Benchmark -bench=BenchmarkWordCount
   ```

   Your parallel implementation should perform better than the provided sequential implementation.
   The TAs will check this during approval.

## High-level Synchronization & Data Race Detection

In this part of the lab we will focus on high-level synchronization techniques using the Go programming language.

The Go language provides a built-in race detector that we will use to identify data races and verify implementations.
A data race occurs when two threads (or goroutines) access a variable concurrently and at least one of these accesses is a write.

We will work on a stack data structure that will be accessed concurrently from several goroutines.
The stack stores values of type `interface{}`, meaning any value type.
The stack interface is shown in Listing 1 and is found in the file `stack_iface.go`.
The interface contains three methods:

* `Size()` returns the current number of items on the stack,
* `Pop() interface{}` pops an item of the stack (`nil` if empty), while
* `Push(value interface{})` pushes an item onto the stack.

Listing 1: Stack interface

```go
type Stack interface {
    Size() int
    Push(value interface{})
    Pop() interface{}
}
```

For this lab we will use the tests defined in `stack_test.go` to verify the different stack implementations.
The tests can be run using the `go test` command. We will run one test at a time.
Running only a specific test can be achieved by supplying the `-run` flag together with a _regular expression_ indicating the test names.
For example, to run only the `TestUnsafeStack` function, use the command:

```console
go test -run TestUnsafeStack
```

There are two type of tests defined for each stack implementation we will be working on.
One test verifies a stack's operations, while the other is meant to test concurrent access using a race detector.
Study the test file for details.

As stated in the introduction, Go includes a built-in data race detector.
Read [Data Race Detector](<http://golang.org/doc/articles/race_detector.html>) for an introduction and usage examples.

### Task: Multiple Choice Questions About Data Races

Answer these multiple choice questions about [Data Races](race_questions.md).

### Task: Implement Thread-Safe Stacks

1. The file `stack_sync.go` is a copy of `stack.go`, but the type is renamed to `SafeStack`.
   Modify this file so that access to the `SafeStack` type is synchronized (can be accessed safely from concurrently running goroutines).
   You can use the `Mutex` type from the `sync` package to achieve this.

2. Verify your implementation by running the `TestSafeStack` test with the data race detector enabled.
   The test should not produce any data race warnings.

   ```console
   go test -race -run TestSafeStack
   ```

3. Go has a built-in high-level API for concurrent programming based on Communicating Sequential Processes (CSP).
   This API promotes synchronization through sending and receiving data via thread-safe channels (as opposed to traditional locking).
   The file `stack_csp.go` contains a `CspStack` type that implements the stack interface in Listing 1 (but the actual method implementations are empty).
   The type also has a constructor function needed for this task.
   Modify this file so that access to the `CspStack` type is synchronized.
   The synchronization should be achieved by using Go’s CSP features (channels and goroutines).

   There is in this case an amount of overhead when using channels to achieve synchronization compared to locking.
   The main point for this task is to give an introduction on how to use channels (CSP) for synchronization.
   This will require some self-study if you are not familiar with Go’s CSP-based concurrent programming capabilities.
   A place to start can be the introduction found [here](http://golang.org/doc/effective_go.html#concurrency).

   Note that you should also ensure that the stack operations are implemented correctly.
   You can verify them by running:

   ```console
   go test -run TestOpsCspStack
   ```

4. Verify your implementation by running the TestCspStack test with the data race detector enabled.
   The test should not produce any data race warnings.

   ```console
   go test -race -run TestCspStack
   ```

## CPU and Memory Profiling

In this part of the lab we will use a technique called profiling to dynamically analyze a program.
Profiling can among other things be used to measure an application’s CPU utilization and memory usage.
Being able to profile applications is very helpful for doing optimizations and is an important part of Systems Programming.
This lab will give a very short introduction to how profiling data can be analyzed.
You may in future lab assignments be required to use profiling to improve and optimize programs.

Profiling for Go can be enabled through the `runtime/pprof` package or by using the testing package’s profiling support.
Profiles can be analyzed and visualized using the `go tool pprof` program.

We will continue to use the stack implementations used in the first part of the lab.
The file `stack_test.go` contains one benchmark for the three different implementations.
Each of them uses the same core stack benchmark defined in the `benchStackOperations(stack Stack)` function.
The stack implementations are not accessed concurrently so that the benchmarks can be kept reasonably deterministic.

Read [Profiling Go Programs](https://blog.golang.org/pprof).
This blog post present a good introduction to Go's profiling abilities.
You should also examine the [testing](http://golang.org/pkg/testing/) package and [testing flags](http://golang.org/cmd/go/#Description_of_testing_flags) for information on how to run the benchmarks, and details about how Go's testing tool easily enables profiling when benchmarking.
Furthermore, we recommend reading [The Go Memory Model](http://golang.org/ref/mem) and [Introducing the Go Race Detector](http://blog.golang.org/race-detector).

### Task: Multiple Choice Questions About CPU and Memory Profiling

Answer these multiple choice questions about [CPU and Memory Profiling](profiling_questions.md).

### Task: Benchmarking and Profiling Go Programs

In this task, you should fill in answers in the provided template: [`benchmark_report.md`](benchmark_report.md).
You can add figures in a directory `fig`, and add markdown links from the benchmark report file, so that the figures display nicely on GitHub's web page.

1. The file `stack_slice.go` contains a stack implementation, `SliceStack`, backed by a slice (dynamic array).
   You will need to adjust this implementation to be synchronized in the exact same way you did for the `SafeStack` type.
   This has to be done to make the benchmark between the three implementations fair and comparable.

2. Run the three stack benchmarks using the following command.

   ```console
   go test -v -run none -bench Benchmark -memprofilerate=1 -benchmem
   ```

   Note that we provide `-run none` in this command, which doesn't match any tests in the `_test.go` file.
   That is we don't run any tests, because we are only interested in the benchmarks, matched by the `-bench Benchmark` flag.
   The command also enables memory allocation statistics by supplying the `-benchmem` flag, and the `-memprofilerate` controls the fraction of memory allocations that are recorded and reported in the memory profile.
   By passing 1 here means all allocations are reported.

   Attach the benchmark output in your [`benchmark_report.md`](benchmark_report.md) and answer the questions.

3. Run the `BenchmarkCspStack` separately and write a CPU profile to file:

   ```console
   go test -v -run none -bench BenchmarkCspStack -cpuprofile=csp-stack.prof
   ```

   Load the CPU profile data with the `pprof` tool.

   ```console
   go tool pprof csp-stack.prof
   ```

   Attach the benchmark and profile output in your [`benchmark_report.md`](benchmark_report.md), and answer the questions related the top ten functions from your CPU profile.

4. Run the `BenchmarkSafeStack` separately and write a memory profile to file:

   ```console
   go test -v -run none -bench BenchmarkSafeStack -memprofile=safe-stack.prof
   ```

   Using the `pprof` tool:

   ```console
   go tool pprof safe-stack.prof
   ```

   Identify the function allocating memory in the `SafeStack` implementation, and list the relevant function to identify the line where the allocations occur.
   Attach the profile output in your [`benchmark_report.md`](benchmark_report.md), and answer the questions related to memory allocations.

5. Install [Graphviz](http://www.graphviz.org/).
   Explore the visualization possibilities offered by `go tool pprof` when analyzing profiling data.
   Use the `pdf` command to produce a call graph:

   ```console
   $ go tool pprof csp-stack.prof
   ...
   (pprof) pdf
   Generating report in profile001.pdf
   (pprof) quit
   ```

   Add the `profile001.pdf` to the `fig/` folder in your group's repository.
   Examine the call graph visualization and answer the questions in the [`benchmark_report.md`](benchmark_report.md).
