// Package stringio provides convenience functions for using strings with
// common io interfaces such as io.Reader and io.Writer, which normally
// use byte slices.
package stringio

import (
	"io"
)

var buffer = make([]byte, 1024)

// Read reads n bytes from the given reader,
// and returns those bytes as a string.
// If the reader provides fewer than n bytes,
// whatever bytes were returned (even if 0
// bytes were returned) will be used to create
// the result string. The returned int and error
// values are taken directly from the return values
// of r.Read().
func Read(r io.Reader, n int) (int, string, error) {
	if cap(buffer) < n {
		newLen := leastPowerOfTwoFunc(n)
		buffer = make([]byte, newLen)
	}
	buffer = buffer[:n]
	n, err := r.Read(buffer)

	buffer = buffer[:n]
	str := string(buffer)
	return n, str, err
}

// Write converts str to a byte slice and
// passes it to w.Write(). The return values
// are passed unmodified from the call to
// w.Write().
func Write(w io.Writer, str string) (int, error) {
	n := len(str)
	if cap(buffer) < n {
		newLen := leastPowerOfTwoFunc(n)
		buffer = make([]byte, newLen)
	}
	buffer = buffer[:n]
	for i := 0; i < n; i++ {
		buffer[i] = str[i]
	}
	return w.Write(buffer)
}

// This will be set to the correct
// function in init()
var leastPowerOfTwoFunc func(int) int = nil

// Assumes 32-bit integers
func leastPowerOfTwoGreaterThan_32B(n int) int {
	// From: http://aggregate.org/MAGIC/#Next%20Largest%20Power%20of%202
	n-- // make sure it's not already a power of 2
	n |= (n >> 1)
	n |= (n >> 2)
	n |= (n >> 4)
	n |= (n >> 8)
	n |= (n >> 16)
	return n + 1
}

// Assumes 64-bit integers
func leastPowerOfTwoGreaterThan_64B(n int) int {
	// From: http://aggregate.org/MAGIC/#Next%20Largest%20Power%20of%202
	n-- // make sure it's not already a power of 2
	n |= (n >> 1)
	n |= (n >> 2)
	n |= (n >> 16)
	n |= (n >> 32)
	return n + 1
}

// In init, check to see what size
// integer values are (ie, int32
// or int64), and set the function
// pointer to leastPowerOfTwoGreaterThan_XXB
// appropriately
func init() {
	// From: http://stackoverflow.com/a/6878625/836390
	var MaxInt int = int(^uint(0) >> 1)
	var Max32BitInt = 0xFFFFFFFF
	if MaxInt > Max32BitInt {
		leastPowerOfTwoFunc = leastPowerOfTwoGreaterThan_64B
	} else {
		leastPowerOfTwoFunc = leastPowerOfTwoGreaterThan_32B
	}
}
