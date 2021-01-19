# Benchmark Report

## CPU and Memory Benchmarks for all Three Stacks

```console
$ go test -v -run none -bench Benchmark -memprofilerate=1 -benchmem
TODO(student)
```

goos: windows
goarch: amd64
pkg: dat320/lab6/stack
BenchmarkSafeStack
BenchmarkSafeStack-8                  45          27332082 ns/op          397954 B/op      19744 allocs/op
BenchmarkSliceStack
BenchmarkSliceStack-8                 99          10159356 ns/op           81264 B/op       9744 allocs/op
BenchmarkCspStack
BenchmarkCspStack-8                   38          29500032 ns/op          100212 B/op       9749 allocs/op
PASS
ok      dat320/lab6/stack       3.763s


1. How much faster than the slowest is the fastest stack?
    - [x] a) 2x-3x
    - [ ] b) 3x-4x
    - [ ] c) 6x-7x
    - [ ] d) 10x-11x

2. Which stack requires the most allocated memory?
    - [ ] a) CspStack
    - [ ] b) SliceStack
    - [x] c) SafeStack
    - [ ] d) UnsafeStack

3. Which stack requires the least amount of allocated memory?
    - [ ] a) CspStack
    - [x] b) SliceStack
    - [ ] c) SafeStack
    - [ ] d) UnsafeStack

## CPU Profile of BenchmarkCspStack

```console
$ go test -v -run none -bench BenchmarkCspStack -cpuprofile=csp-stack.prof
TODO(student)
$ go tool pprof csp-stack.prof
TODO(student)
```
Type: cpu
Time: Oct 19, 2020 at 3:49pm (CEST)
Duration: 2.20s, Total samples = 2.98s (135.32%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top10 
Showing nodes accounting for 2.83s, 94.97% of 2.98s total
Dropped 33 nodes (cum <= 0.01s)
Showing top 10 nodes out of 36
      flat  flat%   sum%        cum   cum%
     1.05s 35.23% 35.23%      1.05s 35.23%  runtime.pthread_cond_signal
     0.86s 28.86% 64.09%      0.86s 28.86%  runtime.pthread_cond_wait
     0.36s 12.08% 76.17%      0.36s 12.08%  runtime.usleep
     0.31s 10.40% 86.58%      0.31s 10.40%  runtime.nanotime1
     0.17s  5.70% 92.28%      0.17s  5.70%  runtime.pthread_mutex_lock
     0.03s  1.01% 93.29%      0.03s  1.01%  runtime.(*waitq).dequeue
     0.02s  0.67% 93.96%      0.02s  0.67%  runtime.(*waitq).dequeueSudoG (inline)
     0.01s  0.34% 94.30%      0.05s  1.68%  dat320/lab6/stack.(*CspStack).Pop
     0.01s  0.34% 94.63%      0.09s  3.02%  dat320/lab6/stack.(*CspStack).run
     0.01s  0.34% 94.97%      1.57s 52.68%  runtime.findrunnable

4. Which function accounts for the most CPU usage?
    - [ ] a) `runtime.pthread_mutex_lock`
    - [ ] b) `dat320/lab6/stack.(*CspStack).run`
    - [ ] c) `runtime.pthread_cond_wait`
    - [x] d) `runtime.pthread_cond_signal`

5. From this top 10 listing, what can you say about the underlying implementation of a CSP-based stack?
    - [x] a) It is implemented using condition variables and locks
    - [ ] b) It is implemented using monitors
    - [ ] c) It is implemented using locks
    - [ ] d) It is implemented using semaphores

## Memory Profile of BenchmarkSafeStack

```console
$ go test -v -run none -bench BenchmarkSafeStack -memprofile=safe-stack.prof -tags solution
TODO(student)
$ go tool pprof safe-stack.prof
TODO(student)
```
Type: alloc_space
Time: Oct 19, 2020 at 3:04pm (CEST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top10
Showing nodes accounting for 531.01MB, 100% of 531.01MB total
      flat  flat%   sum%        cum   cum%
  427.51MB 80.51% 80.51%   427.51MB 80.51%  dat320/lab6/stack.(*SafeStack).Push
  103.50MB 19.49%   100%   531.01MB   100%  dat320/lab6/stack.benchStackOperations
         0     0%   100%   531.01MB   100%  dat320/lab6/stack.BenchmarkSafeStack
         0     0%   100%   531.01MB   100%  testing.(*B).launch
         0     0%   100%   531.01MB   100%  testing.(*B).runN

6. Which function accounts for all memory allocations in the `SafeStack` implementation?
    - [ ] a) `Size`
    - [ ] b) `NewSafeStack`
    - [x] c) `Push`
    - [ ] d) `Pop`

7. Which line in `SafeStack` does the actual memory allocation?
    - [ ] a) `type SafeStack struct {`
    - [x] b) `ss.top = &Element{value, ss.top}`
    - [ ] c) `value, ss.top = ss.top.value, ss.top.next`
    - [ ] d) `top  *Element`

## Visualizing the SafeStack Call Graph

Examine the call graph visualization and answer the questions below.

8. What is the root node of the call graph that ends with `runtime.pthread_cond_wait`?
    - [ ] a) `runtime.park_m`
    - [ ] b) `runtime.schedule`
    - [ ] c) `runtime.mstart`
    - [x] d) `runtime.mcall`

9. Which of the following consume the most CPU time?
    - [ ] a) `runtime.park_m`
    - [x] b) `runtime.usleep`
    - [ ] c) `runtime.mstart`
    - [ ] d) `runtime.lock`
