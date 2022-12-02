package caller

import "runtime"

// pc
// depth= 0: the caller position
// depth=-1: this function line 10
func PC(depth int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(depth+2, pcs[:])
	return pcs[0]
}
