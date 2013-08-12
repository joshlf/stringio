// Copyright 2013 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stringio

import (
	"strings"
	// "bufio"
	"bytes"
	"sync"
	"testing"
)

// Test a highly-parallel approach to make sure
// that the buffer blocking strategy works
// (ie, that we don't use the same buffer as
// scratch space in multiple goroutines simultaneously)
const threads = 1000

var testStrings = []string{
	"abc 123", // generic test
	"日本語",     // test unicode (runes would be a likely stumbling block)
	// This string's length is 2049. It is meant to test the buffer growing machinery.
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
}

func TestRead(t *testing.T) {
	w := new(sync.WaitGroup)
	for i := 0; i < threads; i++ {
		w.Add(1)
		testRead(t, w)
	}
	w.Wait()
}

func testRead(t *testing.T, w *sync.WaitGroup) {
	for _, s := range testStrings {
		testReadHelper(s, t)
	}
	w.Done()
}

func testReadHelper(str string, t *testing.T) {
	rdr := strings.NewReader(str)
	n := len(str)
	n, strPrime, err := Read(rdr, n)
	if err != nil {
		t.Errorf("Read returned error: %v\n", err)
	} else if n != len(str) {
		t.Errorf("Read read %n bytes; should have read %n\n", n, len(str))
	} else if strPrime != str {
		t.Errorf("Read returned the wrong string: \"%s\" instead of \"%s\"", strPrime, str)
	}
}

func TestWrite(t *testing.T) {
	w := new(sync.WaitGroup)
	for i := 0; i < threads; i++ {
		w.Add(1)
		go testWrite(t, w)
	}
	w.Wait()
}

func testWrite(t *testing.T, w *sync.WaitGroup) {
	for _, s := range testStrings {
		testWriteHelper(s, t)
	}
	w.Done()
}

func testWriteHelper(str string, t *testing.T) {
	// Capacity of slice must be 0
	// otherwise leading bytes will
	// be counted as part of the buffer
	// and will not be overwritten.
	wrtr := bytes.NewBuffer(make([]byte, 0, len(str)))
	n, err := Write(wrtr, str)
	if err != nil {
		t.Errorf("Write returned error: %v\n", err)
	} else if n != len(str) {
		t.Errorf("Write wrote %n bytes; should have written %n\n", n, len(str))
	} else {
		strPrime := wrtr.String()
		if strPrime != str {
			t.Errorf("Write wrote the wrong string: \"%s\" instead of \"%s\"", strPrime, str)
		}
	}
}

func TestReadAt(t *testing.T) {
	w := new(sync.WaitGroup)
	for i := 0; i < threads; i++ {
		w.Add(1)
		go testReadAt(t, w)
	}
	w.Wait()
}

func testReadAt(t *testing.T, w *sync.WaitGroup) {
	for _, s := range testStrings {
		for i := int64(0); i < 1024; i++ {
			testReadAtHelper(s, i, t)
		}
	}
	w.Done()
}

func testReadAtHelper(str string, off int64, t *testing.T) {
	b := make([]byte, int(off)+len(str))
	for i := int(off); i < len(b); i++ {
		b[i] = str[i-int(off)]
	}
	rdr := strings.NewReader(string(b))
	n := len(str)
	n, strPrime, err := ReadAt(rdr, n, off)
	if err != nil {
		t.Errorf("Read returned error: %v\n", err)
	} else if n != len(str) {
		t.Errorf("Read read %n bytes; should have read %n\n", n, len(str))
	} else if strPrime != str {
		t.Errorf("Read returned the wrong string: \"%s\" instead of \"%s\"", strPrime, str)
	}
}

func TestLeastPowerGreaterThan(t *testing.T) {
	testLeastPowerGreaterThan(0, 0, t)

	// This test should take around 2 seconds.
	// "i += 2" to reduce runtime; it
	// is still a sufficient test.
	for i := 0; i < 32; i += 2 {
		target := 1 << uint(i-1)
		start := (target >> 1) + 1
		for i := start; i <= target; i++ {
			testLeastPowerGreaterThan(target, leastPowerOfTwoFunc(i), t)
		}
	}

	if !_32bit {
		maxInt := int((^uint(0)) - (uint(1) << 63))
		testLeastPowerGreaterThan(maxInt+1, leastPowerOfTwoFunc(maxInt), t)
	}
}

// Returns whether or not the test was successful
func testLeastPowerGreaterThan(i, j int, t *testing.T) bool {
	if i != j {
		t.Errorf("Expected %d; got %d\n", i, j)
		return false
	}
	return true
}

var _32bit bool

func init() {
	// From: http://stackoverflow.com/a/6878625/836390
	var MaxInt int = int(^uint(0) >> 1)
	var Max32BitInt = 0xFFFFFFFF
	if MaxInt > Max32BitInt {
		_32bit = false
	} else {
		_32bit = true
	}
}
