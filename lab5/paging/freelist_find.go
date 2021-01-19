package paging

// findFreeFrames returns indices for n free frames.
// If there are not enough free frames available, an error is returned.
func (fl *freeList) findFreeFrames(n int) ([]int, error) {
	if n > fl.numFreeFrames {
		return nil, errOutOfMemory
	}

	freeFrames := []int{}
	for i, entry := range fl.freeList {
		if entry {
			freeFrames = append(freeFrames, i)
		}

		if len(freeFrames) == n {
			// We have all the frames we need so break the loop and return the frames
			break
		}
	}
	return freeFrames, nil
}
