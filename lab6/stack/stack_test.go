package stack

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestUnsafeStack(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	unsafeStack := new(UnsafeStack)
	fmt.Println("Unsafe Stack Test")
	testConcurrentStackAccess(unsafeStack)
}

func TestSafeStack(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	safeStack := new(SafeStack)
	fmt.Println("Safe Stack Test")
	testConcurrentStackAccess(safeStack)
}

func TestCspStack(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cspStack := NewCspStack()
	fmt.Println("CSP Stack Test")
	testConcurrentStackAccess(cspStack)
}

func TestSliceStack(t *testing.T) {
	sliceStack := NewSliceStack()
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Slice Stack Test")
	testConcurrentStackAccess(sliceStack)
}

func TestOpsUnsafeStack(t *testing.T) {
	fmt.Println("Test operations UnsafeStack")
	unsafeStack := new(UnsafeStack)
	testStackOperations(unsafeStack, t)
}

func TestOpsSafeStack(t *testing.T) {
	fmt.Println("Test operations SafeStack")
	safeStack := new(SafeStack)
	testStackOperations(safeStack, t)
}

func TestOpsCspStack(t *testing.T) {
	fmt.Println("Test operations CspStack")
	cspStack := NewCspStack()
	testStackOperations(cspStack, t)
}

func TestOpsSliceStack(t *testing.T) {
	fmt.Println("Test operations SliceStack")
	sliceStack := NewSliceStack()
	testStackOperations(sliceStack, t)
}

func TestOpsAllStacks(t *testing.T) {
	fmt.Println("Test operations all stacks")
	TestOpsUnsafeStack(t)
	TestOpsSafeStack(t)
	TestOpsCspStack(t)
	TestOpsSliceStack(t)
}

func BenchmarkSafeStack(b *testing.B) {
	safeStack := new(SafeStack)
	for i := 0; i < b.N; i++ {
		benchStackOperations(safeStack)
	}
}

func BenchmarkSliceStack(b *testing.B) {
	sliceStack := NewSliceStack()
	for i := 0; i < b.N; i++ {
		benchStackOperations(sliceStack)
	}
}

func BenchmarkCspStack(b *testing.B) {
	sliceStack := NewCspStack()
	for i := 0; i < b.N; i++ {
		benchStackOperations(sliceStack)
	}
}

const (
	numGoroutines = 4
	numOperations = 10
)

const (
	Len = iota
	Push
	Pop
)

func testConcurrentStackAccess(stack Stack) {
	rand.Seed(time.Now().Unix())
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			for j := 0; j < numOperations; j++ {
				op := rand.Intn(3)
				switch op {
				case Len:
					// fmt.Printf("#%d-%d: Size() was %d\n", i, j, stack.Size())

					stack.Size()
				case Push:
					// data := "Data" + strconv.Itoa(i) + strconv.Itoa(j)
					// fmt.Printf("#%d-%d: Push() with value %v\n", i, j, data)

					stack.Push("Data" + strconv.Itoa(i) + strconv.Itoa(j))
				case Pop:
					_ = stack.Pop()

					// fmt.Printf("#%d-%d: Pop() gave value %v\n", i, j, value)
				}
			}

			defer wg.Done()
		}(i)
	}

	wg.Wait()
}

func testStackOperations(stack Stack, t *testing.T) {
	var length int
	var wantLength int
	var testName string

	// Initial Stack Test
	testName = "Initial Stack Test"
	wantLength = 0
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	// Pushed One Test
	testName = "Pushed One Test"
	stack.Push("Item1")
	wantLength = 1
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	item1 := stack.Pop()
	if item1 != "Item1" {
		t.Errorf("\n%s\nAction: Pop()\nWant: Item1\nGot: %v", testName, item1)
	}
	wantLength = 0
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	// Pushed Three Test
	testName = "Pushed Three Test"
	stack.Push("Item2")
	stack.Push(3)
	stack.Push(4.0001)
	wantLength = 3
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	item4 := stack.Pop()
	if item4 != 4.0001 {
		t.Errorf("\n%s\nAction: Pop()\nWant: 4.0001\nGot: %v", testName, item4)
	}
	wantLength = 2
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	item3 := stack.Pop()
	if item3 != 3 {
		t.Errorf("\n%s\nAction: Pop()\nWant: 3\nGot: %v", testName, item3)
	}
	wantLength = 1
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	item2 := stack.Pop()
	if item2 != "Item2" {
		t.Errorf("\n%s\nAction: Pop()\nWant: Item2\nGot: %v", testName, item2)
	}
	wantLength = 0
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	item5 := stack.Pop()
	if item5 != nil {
		t.Errorf("\n%s\nAction: Pop()\nWant: <nil>\nGot: %v", testName, item5)
	}
	wantLength = 0
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	// Stack Slice Allocation Test
	testName = "Stack Slice Allocation Test"
	size := 200
	for i := 0; i < size; i++ {
		stack.Push(i)
	}
	wantLength = size
	if length = stack.Size(); length != wantLength {
		t.Errorf("\n%s\nAction: Size()\nWant: %d\nGot: %d", testName, wantLength, length)
	}

	for j := size - 1; j >= 0; j-- {
		if x := stack.Pop(); x != j {
			t.Errorf("\n%s\nAction: Pop()\nWant: %d\nGot: %v", testName, j, x)
			break
		}
	}
}

func benchStackOperations(stack Stack) {
	const numOps = 10000
	for i := 0; i < numOps; i++ {
		stack.Push(i)
	}
	for j := 0; j < numOps; j++ {
		stack.Pop()
	}
}
