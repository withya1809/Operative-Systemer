package paging

// MMU is the structure for the simulated memory management unit.
type MMU struct {
	frames    [][]byte           // contains memory content in form of frames[frameIndex][offset]
	freeList                     // tracks free physical frames
	processes map[int]*PageTable // contains page table for each process (key=pid)
}

// OffsetLookupTable gives the bit mask corresponding to a virtual address's offset of length n,
// where n is the table index. This table can be used to find the offset mask needed to extract
// the offset from a virtual address. It supports up to 32-bit wide offset masks.
//
// OffsetLookupTable[0] --> 0000 ... 0000
// OffsetLookupTable[1] --> 0000 ... 0001
// OffsetLookupTable[2] --> 0000 ... 0011
// OffsetLookupTable[3] --> 0000 ... 0111
// OffsetLookupTable[8] --> 0000 ... 1111 1111
// etc.
var OffsetLookupTable = []int{
	// 0000, 0001, 0011, 0111, 1111, etc.
	0x0000000, 0x00000001, 0x00000003, 0x00000007,
	0x000000f, 0x0000001f, 0x0000003f, 0x0000007f,
	0x00000ff, 0x000001ff, 0x000003ff, 0x000007ff,
	0x0000fff, 0x00001fff, 0x00003fff, 0x00007fff,
	0x000ffff, 0x0001ffff, 0x0003ffff, 0x0007ffff,
	0x00fffff, 0x001fffff, 0x003fffff, 0x007fffff,
	0x0ffffff, 0x01ffffff, 0x03ffffff, 0x07ffffff,
	0xfffffff, 0x1fffffff, 0x3fffffff, 0x7fffffff, 0xffffffff,
}

// NewMMU creates a new MMU with a memory of memSize bytes.
// memSize should be >= 1 and a multiple of frameSize.
func NewMMU(memSize, frameSize int) *MMU {
	numFrames := memSize / frameSize
	// Initializes a slice of booleans for the frames. All frames are available in the start so all entries in the slice are true
	// boolList := make([]bool, numFrames)
	// for i := 0; i < numFrames; i++ {
	// 	boolList[i] = true
	// }
	// Initializes the 2d slice of bytes. [numFrames][numBytesPerFrame]byte. Stores a zero value in all frames
	byteSlice := make([][]byte, numFrames)
	for i := range byteSlice {
		tempSlice := make([]byte, frameSize)
		for j := range tempSlice {
			tempSlice[j] = 0x00
		}
		byteSlice[i] = tempSlice
	}

	return &MMU{
		frames:    byteSlice,
		freeList:  newFreeList(numFrames),
		processes: make(map[int]*PageTable),
	}
}

// Alloc allocates n bytes of memory for process pid.
// The allocated memory is added to the process's page table.
// The process is given a page table if it doesn't already have one,
// unless an out of memory error occurred.
func (mmu *MMU) Alloc(pid, n int) error {
	// Suggested approach:
	// - calculate #frames needed to allocate n bytes, error if not enough free frames
	// - if process pid has no page table, create one for it
	// - determine which frames to allocate to the process
	// - add the frames to the process's (identified by pid) page table and
	// - update the free list

	if n < 1 {
		return errNothingToAllocate
	}

	// Find requested mount of frames
	numFrames := 0
	// BytesPerFrame = len(mmu.frames[0]) as long as the length is constant
	bytesPerFrame := len(mmu.frames[0])
	for n > 0 {
		n -= bytesPerFrame
		numFrames++
	}

	physicalFrames, err := mmu.freeList.findFreeFrames(numFrames)
	if err != nil {
		return err
	}

	var pageTable *PageTable
	// Get the process unique page table
	if mmu.processes[pid] == nil {
		pageTable = &PageTable{
			frameIndices: []int{},
		}
		mmu.processes[pid] = pageTable
	} else {
		pageTable = mmu.processes[pid]
	}

	pageTable.Append(physicalFrames)
	// uppdates the free list
	err = mmu.freeList.removeFrames(physicalFrames)
	if err != nil {
		return err
	}
	return nil
}

//Withya
// Write writes content to the given process's address space starting at virtualAddress.
func (mmu *MMU) Write(pid, virtualAddress int, content []byte) error {
	// - check valid pid (must have a page table)
	_, err := mmu.getPageTable(pid)
	if err != nil {
		return err
	}
	// - translate the virtual address
	vpn, offset, r := mmu.translateAndCheck(pid, virtualAddress)

	if r != nil { //illegal address
		return r
	}

	frameSize := len(mmu.frames[0])
	bytesLeft := (frameSize - offset) + (mmu.processes[pid].Len()-vpn-1)*frameSize //resterende bytes i nåværende minne fra start_Addressen

	// - check if the memory must be extended in order to write the content
	// - attempt to allocate more memory if necessary to complete the write

	if len(content) > bytesLeft { //trenger mer bytes enn det som er igjen i current frame
		n := len(content) - bytesLeft // finner resterende bytes som er igjen. Må allokere mer minne

		alloc_err := mmu.Alloc(pid, n)
		if alloc_err != nil {
			return errFreeOutOfBounds
		}

	}

	// - sequentially write content into the known-to-be-valid address space
	content_num := 0

	for i := vpn; true; i++ { //så lenge den holder seg til samme vpn

		physicalFrameIndex, errr := mmu.processes[pid].Lookup(i) // finner physical address av current vpn
		if errr != nil {
			return errr
		}

		for o := offset; o < frameSize; o++ { //øker offset for hver iterasjon
			mmu.frames[physicalFrameIndex][o] = content[content_num] // setter content til current offset
			content_num += 1                                         //går til neste byte i content lista

			if content_num == len(content) {
				return nil //break loop, no more content to write
			}
		}

		offset = 0 //hvis offset har blitt lenger enn frameSize, må den nullstilles

	}
	return nil

}

// Read returns content of size n bytes from the given process's address space starting at virtualAddress.
func (mmu *MMU) Read(pid, virtualAddress, n int) (content []byte, err error) {
	// Suggested approach:
	// - check valid pid (must have a page table)
	// - translate the virtual address
	// - (optional) determine if it's possible to read the requested number
	//   of bytes before starting to read the memory content
	// - read and return the requested memory content
	if n < 1 {
		return nil, errNothingToRead
	}
	pageTable, err := mmu.getPageTable(pid)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		vpn, currentByte, err := mmu.translateAndCheck(pid, virtualAddress)
		if err != nil {
			return nil, err
		}
		frameNumber, err := pageTable.Lookup(vpn)
		if err != nil {
			return nil, err
		}
		content = append(content, mmu.frames[frameNumber][currentByte])
		virtualAddress = virtualAddress + 1
	}
	return content, nil
}

// Free is called by a process's Free() function to free some of its allocated memory.
func (mmu *MMU) Free(pid, n int) error {
	// Suggested approach:

	// - check valid pid (must have a page table)
	pageTable, err := mmu.getPageTable(pid)
	if err != nil {
		return err
	}

	// - check if there are at least n entries in the page table of pid
	if pageTable.Len() < n {
		return errFreeOutOfBounds
	}
	// Frees n frames from the pageTable
	// vpns is a list of the physical frames freed
	physicalFramesFreed, err := pageTable.Free(n)
	if err != nil {
		return err
	}

	// - set all the bytes in the freed memory to the value 0
	offset := 0
	frameSize := len(mmu.frames[0])

	for _, i := range physicalFramesFreed {
		for o := offset; o < frameSize; o++ {
			mmu.frames[i][o] = 0
		}
		offset = 0
	}

	// - re-add the freed frames to the free list
	mmu.freeList.addFrames(physicalFramesFreed)

	return nil
}

//Withya
// extract returns the virtual page number and offset for the given virtual address,
// and the number of bits in the offset n.
func extract(virtualAddress, n int) (vpn, offset int) {
	// the Virtual Addresses section of the README.
	// The procedure is described in detail in Chapter 18.1 of the textbook.
	// HINT: It can be solved quite easily with bitwise operators.
	// (see https://golang.org/ref/spec#Arithmetic_operators ).
	// You might also find the provided log2 function and the OffsetLookupTable
	// table useful for this purpose.

	offsetMask := OffsetLookupTable[n]       //offsetmask
	vA_offset := virtualAddress & offsetMask //AND operation mellom VA og OffsetMask gir oss Offset

	vA_vpn := virtualAddress >> n

	return vA_vpn, vA_offset

}

// translateAndCheck returns the virtual page number and offset for the given virtual address.
// If the virtual address is invalid for process pid, an error is returned.
func (mmu *MMU) translateAndCheck(pid, virtualAddress int) (vpn, offset int, err error) {
	// the Virtual Addresses section of the README.
	// The procedure is described in detail in Chapter 18.1 of the textbook.
	// It is expected that this method calls the extract function above
	// to compute the VPN and offset to be returned from this function after
	// checking that the process has access to the returned VPN.
	// You might also find the provided log2 function useful to calculate one
	// of the inputs to the extract function.

	frameSize := len(mmu.frames[0]) //finner framSize
	n := log2(frameSize)            //antall bits for offset
	vA_vpn, vA_offset := extract(virtualAddress, n)

	pageTable, r := mmu.getPageTable(pid)
	if r != nil {
		return 0, 0, r
	}

	_, errr := pageTable.Lookup(vA_vpn)
	if errr != nil {
		return 0, 0, errr
	}

	return vA_vpn, vA_offset, nil

}

func (mmu *MMU) getPageTable(pid int) (pageTable *PageTable, err error) {
	if pageTable, ok := mmu.processes[pid]; ok {
		return pageTable, nil
	}
	return nil, errInvalidProcess
}

// log2 calculates m given n = 2^m.
func log2(n int) int {
	exp := 0
	for {
		if n%2 == 0 && n > 0 {
			exp++
		} else {
			return exp
		}
		n /= 2
	}
}
