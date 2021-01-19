package stack

// CspStack is a struct with methods needed to implement the Stack interface.
type CspStack struct {
	pushChan      chan interface{}
	popStartChan  chan interface{}
	popReturnChan chan interface{}
	lenChan       chan interface{}
	lenReturnChan chan int
	stack         []interface{}
}

// NewCspStack returns an empty CspStack.
func NewCspStack() *CspStack {
	cspStack := &CspStack{
		pushChan:      make(chan interface{}),
		popReturnChan: make(chan interface{}),
		popStartChan:  make(chan interface{}),
		lenChan:       make(chan interface{}),
		lenReturnChan: make(chan int),
		stack:         []interface{}{},
	}
	go cspStack.run()
	return cspStack
}

// Size returns the size of the stack.
func (cs *CspStack) Size() int {
	// Signal that we want to know the size
	cs.lenChan <- true
	return <-cs.lenReturnChan
}

// Push pushes value onto the stack.
func (cs *CspStack) Push(value interface{}) {
	cs.pushChan <- value
}

// Pop pops the value at the top of the stack and returns it.
func (cs *CspStack) Pop() (value interface{}) {
	// Signal that we want to pop of the stack
	cs.popStartChan <- true
	return <-cs.popReturnChan
}

func (cs *CspStack) run() {
	var value interface{}
	for {
		select {
		case <-cs.lenChan:
			cs.lenReturnChan <- len(cs.stack)
		case value = <-cs.pushChan:
			cs.stack = append(cs.stack, value)
		case <-cs.popStartChan:
			if len(cs.stack) > 0 {
				value = cs.stack[len(cs.stack)-1]
				cs.stack = cs.stack[:len(cs.stack)-1]
				cs.popReturnChan <- value
			} else {
				// If there are no elements to pop: return nil
				cs.popReturnChan <- nil
			}

		}
	}
}
