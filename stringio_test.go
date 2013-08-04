// Copyright 2013 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stringio

import (
	"strings"
	// "bufio"
	"bytes"
	"testing"
)

var testStrings = []string{
	"abc 123", // generic test
	"日本語",     // test unicode (runes would be a likely stumbling block)

}

func TestRead(t *testing.T) {
	for _, s := range testStrings {
		testRead(s, t)
	}
}

func testRead(str string, t *testing.T) {
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
	for _, s := range testStrings {
		testWrite(s, t)
	}
}

func testWrite(str string, t *testing.T) {
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

func TestLeastPowerGreaterThan(t *testing.T) {
	testLeastPowerGreaterThan(0, 0, t)

	// This test should take around 2 seconds.
	// Increase by 2 to reduce runtime; it
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
