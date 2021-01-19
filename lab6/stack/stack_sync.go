package stack

import (
	"sync"
)

//Withya (1&2)

// SafeStack holds the top element of the stack and its size.
type SafeStack struct {
	top  *Element
	size int
}

var mutex sync.Mutex

// Size returns the size of the stack.
func (ss *SafeStack) Size() int {

	mutex.Lock()
	defer mutex.Unlock()
	size := ss.size
	return size

}

// Push pushes value onto the stack.
func (ss *SafeStack) Push(value interface{}) {
	mutex.Lock()
	ss.top = &Element{value, ss.top}
	ss.size++
	mutex.Unlock()
}

// Pop pops the value at the top of the stack and returns it.
func (ss *SafeStack) Pop() (value interface{}) {

	mutex.Lock()
	defer mutex.Unlock()
	if ss.size > 0 {
		value, ss.top = ss.top.value, ss.top.next
		ss.size--
		return
	}
	//mutex.Unlock()

	return nil
}
