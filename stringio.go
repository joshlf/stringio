// Copyright 2013 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package stringio provides convenience functions for using strings with
// common io interfaces such as io.Reader and io.Writer, which normally
// use byte slices.
package stringio

import (
	"github.com/joshlf13/sync"
	"io"
)

// The strategy is to use this buffer as scratch space
// when possible. However, we don't want two goroutines
// using it as scratch space simultaneously. Thus, the
// mutex is used to make sure that only one goroutine
// at a time uses the buffer. However, we don't want
// to force only one call to this package to happen
// at a time, so if the lock can't be acquired, simply
// go with the more expensive solution and allocate a
// temporary buffer. The reasoning here is that in most
// applications, these lockouts will be rare, so while
// they are expensive when they do happen, the overall
// cost is relatively small.
var (
	buffer = make([]byte, 1024)
	mutex  sync.Mutex
)

// Read reads n bytes from the given io.Reader,
// and returns those bytes as a string.
// If the io.Reader provides fewer than n bytes,
// whatever bytes were returned (even if 0
// bytes were returned) will be used to create
// the result string. The returned int and error
// values are taken directly from the return values
// of r.Read().
func Read(r io.Reader, n int) (int, string, error) {
	if mutex.LockIfPossible() {
		defer mutex.Unlock()
		return readWithBuffer(r, n)
	}
	return readWithoutBuffer(r, n)
}

func readWithBuffer(r io.Reader, n int) (int, string, error) {
	extendAndSliceBuffer(n)
	n, err := r.Read(buffer)

	buffer = buffer[:n]
	str := string(buffer)
	return n, str, err
}

func readWithoutBuffer(r io.Reader, n int) (int, string, error) {
	buf := make([]byte, n)
	n, err := r.Read(buf)

	buf = buf[:n]
	str := string(buf)
	return n, str, err
}

// Write converts str to a byte slice and
// passes it to w.Write(). The return values
// are passed unmodified from the call to
// w.Write().
//
// Write is identical in behavior to
// io.WriteString, although it will usually
// be somewhat more performant since it
// avoids reallocating len(str) bytes
// if possible.
func Write(w io.Writer, str string) (int, error) {
	if mutex.LockIfPossible() {
		defer mutex.Unlock()
		return writeWithBuffer(w, str)
	}
	return writeWithoutBuffer(w, str)
}

func writeWithBuffer(w io.Writer, str string) (int, error) {
	// We could in theory take a page out of
	// io.WriteString's book at check if the
	// given io.Writer has a WriteString method.
	// However, such a method would probably
	// reallocate just like io.WriteString does,
	// so avoiding that will probably more efficient.
	//
	// (See http://golang.org/src/pkg/io/io.go?s=9270:9325#L254
	// for the source of io.WriteString)
	n := len(str)
	extendAndSliceBuffer(n)
	for i := 0; i < n; i++ {
		buffer[i] = str[i]
	}
	return w.Write(buffer)
}

type stringWriter interface {
	WriteString(string) (int, error)
}

func writeWithoutBuffer(w io.Writer, str string) (int, error) {
	// Since we don't have the buffer available,
	// a copy is going to have to happen. Thus,
	// even if sw.WriteString does a copy, it
	// will be no more expensive than if we had
	// done it ourselves.
	if sw, ok := w.(stringWriter); ok {
		return sw.WriteString(str)
	}
	return w.Write([]byte(str))
}

// ReadAt reads n bytes from the given io.ReaderAt,
// and returns those bytes as a string. If the
// io.ReaderAt provides fewer than n bytes, whatever
// bytes were provided (even if 0 bytes were provided)
// will be used to create the result string. The
// returned int and error values are taken directly
// from the return values of r.ReadAt()
func ReadAt(r io.ReaderAt, n int, off int64) (int, string, error) {
	if mutex.LockIfPossible() {
		defer mutex.Unlock()
		return readAtWithBuffer(r, n, off)
	}
	return readAtWithoutBuffer(r, n, off)
}

func readAtWithBuffer(r io.ReaderAt, n int, off int64) (int, string, error) {
	extendAndSliceBuffer(n)
	n, err := r.ReadAt(buffer, off)

	buffer = buffer[:n]
	str := string(buffer)
	return n, str, err
}

func readAtWithoutBuffer(r io.ReaderAt, n int, off int64) (int, string, error) {
	buf := make([]byte, n)
	n, err := r.ReadAt(buf, off)

	buf = buf[:n]
	str := string(buf)
	return n, str, err
}

// WriteAt converts str to a byte slice and
// passes it to w.WriteAt() along with off.
// The return values are passed unmodified
// from the call to w.WriteAt().
func WriteAt(w io.WriterAt, str string, off int64) (int, error) {
	if mutex.LockIfPossible() {
		defer mutex.Unlock()
		return writeAtWithBuffer(w, str, off)
	}
	return writeAtWithoutBuffer(w, str, off)
}

func writeAtWithBuffer(w io.WriterAt, str string, off int64) (int, error) {
	n := len(str)
	extendAndSliceBuffer(n)
	for i := 0; i < n; i++ {
		buffer[i] = str[i]
	}
	return w.WriteAt(buffer, off)
}

func writeAtWithoutBuffer(w io.WriterAt, str string, off int64) (int, error) {
	return w.WriteAt([]byte(str), off)
}

// If cap(buffer) < n, reallocate
// the buffer so that its capacity
// is the least power of two larger
// than n. Slice buffer so that its
// length is equal to n.
func extendAndSliceBuffer(n int) {
	if cap(buffer) < n {
		buffer = make([]byte, leastPowerOfTwoFunc(n))
	}
	buffer = buffer[:n]
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
	n |= (n >> 4)
	n |= (n >> 8)
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
