package stringio

import (
	"strings"
	// "bufio"
	"bytes"
	"testing"
)

var testStrings = []string{
	"abc 123",
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
