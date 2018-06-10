package goobj

import (
	"bufio"
	"strings"
	"testing"
)

func TestParser_skipHeader(t *testing.T) {
	for i, testData := range []struct {
		in       string
		expected int64
	}{
		{in: "\x00\x00go19ld", expected: 8},
		{in: "\x00\x00\x00go19ld", expected: 9},
		{in: "\x00\x00go19l\x00\x00go19ld", expected: 15},
	} {
		p := newParser(bufio.NewReader(strings.NewReader(testData.in)))
		p.skipHeader()
		if p.Err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, p.Err)
		}
		if p.reader.numReadBytes != testData.expected {
			t.Errorf("[%d] the number of read bytes should be %d, but %d", i, testData.expected, p.reader.numReadBytes)
		}
	}
}

func TestParser_skipHeader_EmptyInput(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("")))
	p.skipHeader()
	if p.Err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_skipHeader_HeaderNotFound(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x00\x00\x00\x00\x00\x00\x00\x00")))
	p.skipHeader()
	if p.Err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_checkVersion(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x01")))
	p.checkVersion()
	if p.Err != nil {
		t.Errorf("error should be nil")
	}
	if p.reader.numReadBytes != 1 {
		t.Errorf("the number of read bytes should be 1, but %d", p.reader.numReadBytes)
	}
}

func TestParser_checkVersion_NotSupportedVersion(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x00")))
	p.checkVersion()
	if p.Err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_skipDependencies(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x01\x00")))
	p.skipDependencies()
	if p.Err != nil {
		t.Errorf("error should be nil")
	}
	if p.reader.numReadBytes != 2 {
		t.Errorf("the number of read bytes should be 2, but %d", p.reader.numReadBytes)
	}
}

func TestParser_skipDependencies_EmptyInput(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("")))
	p.skipDependencies()
	if p.Err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_readReferences(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\xfe\x02a\x02\xfe\x02b\x02\xff")))
	p.readReferences()
	if len(p.symbolReferences) != 2 {
		t.Errorf("the number of symbolReferences should be 2, but %d", len(p.symbolReferences))
	}
	if p.symbolReferences[0].Name != "a" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
	if p.symbolReferences[1].Name != "b" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
	if p.Err != nil {
		t.Errorf("error should be nil")
	}
	if p.reader.numReadBytes != 9 {
		t.Errorf("the number of read bytes should be 8, but %d", p.reader.numReadBytes)
	}
}

func TestParser_readReference(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x02a\x02")))
	p.readReference()
	if len(p.symbolReferences) != 1 {
		t.Errorf("the number of symbolReferences should be 1, but %d", len(p.symbolReferences))
	}
	if p.symbolReferences[0].Name != "a" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
	if p.Err != nil {
		t.Errorf("error should be nil")
	}
	if p.reader.numReadBytes != 3 {
		t.Errorf("the number of read bytes should be 3, but %d", p.reader.numReadBytes)
	}
}

func TestReaderWithCounter_readVarint(t *testing.T) {
	for i, testData := range []struct {
		in       string
		expected int64
	}{
		{in: "\x00", expected: 0},
		{in: "\x01", expected: -1},
		{in: "\x02", expected: 1},
		{in: "\x80\x01", expected: 64},
		{in: "\x81\x01", expected: -65},
	} {
		reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader(testData.in))}
		actual, err := reader.readVarint()
		if actual != testData.expected {
			t.Errorf("[%d] the value should be %d, but %d", i, testData.expected, actual)
		}
		if err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, err)
		}
		if reader.numReadBytes != int64(len(testData.in)) {
			t.Errorf("[%d] the number of read bytes should be %d, but %d", i, len(testData.in), reader.numReadBytes)
		}
	}
}

func TestReaderWithCounter_readVarint_Error(t *testing.T) {
	reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader(""))}
	_, err := reader.readVarint()
	if err == nil {
		t.Errorf("error is not recorded")
	}
}

func TestReaderWithCounter_readString(t *testing.T) {
	for i, testData := range []struct {
		in       string
		expected string
	}{
		{in: "\x00", expected: ""},
		{in: "\x02a", expected: "a"},
		{in: "\x04ab", expected: "ab"},
	} {
		reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader(testData.in))}
		actual, err := reader.readString()
		if actual != testData.expected {
			t.Errorf("[%d] the value should be %d, but %d", i, testData.expected, actual)
		}
		if err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, err)
		}
		if reader.numReadBytes != int64(len(testData.expected)+1) {
			t.Errorf("[%d] the number of read bytes should be %d, but %d", i, len(testData.in), reader.numReadBytes)
		}
	}
}

func TestReaderWithCounter_readString_TooShortString(t *testing.T) {
	reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader("\x02"))}
	_, err := reader.readString()
	if err == nil {
		t.Errorf("error should not be nil")
	}
}
