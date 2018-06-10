package goobj

import (
	"bufio"
	"bytes"
	"reflect"
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
		if err := p.skipHeader(); err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, err)
		}
	}
}

func TestParser_skipHeader_EmptyInput(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("")))
	if err := p.skipHeader(); err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_skipHeader_HeaderNotFound(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x00\x00\x00\x00\x00\x00\x00\x00")))
	if err := p.skipHeader(); err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_checkVersion(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x01")))
	if err := p.checkVersion(); err != nil {
		t.Errorf("error should be nil")
	}
}

func TestParser_checkVersion_NotSupportedVersion(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x00")))
	if err := p.checkVersion(); err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_skipDependencies(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x01\x00")))
	if err := p.skipDependencies(); err != nil {
		t.Errorf("error should be nil")
	}
}

func TestParser_skipDependencies_EmptyInput(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("")))
	if err := p.skipDependencies(); err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_readReferences(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\xfe\x02a\x02\xfe\x02b\x02\xff")))
	err := p.readReferences()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.symbolReferences) != 2 {
		t.Errorf("the number of symbolReferences should be 2, but %d", len(p.symbolReferences))
	}
	if p.symbolReferences[0].Name != "a" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
	if p.symbolReferences[1].Name != "b" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
}

func TestParser_readReference(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x02a\x02")))
	err := p.readReference()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.symbolReferences) != 1 {
		t.Errorf("the number of symbolReferences should be 1, but %d", len(p.symbolReferences))
	}
	if p.symbolReferences[0].Name != "a" || p.symbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.symbolReferences[0])
	}
}

func TestParser_readData(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x02\x00\x00\x00\x00\x00a")))
	err := p.readData()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if !reflect.DeepEqual([]byte("a"), p.Data) {
		t.Errorf("the data should be a, but %s", string(p.Data))
	}
}

func TestParser_readData_128KBData(t *testing.T) {
	dataLength := "\x80\x80\x10" // 128KB
	data := strings.Repeat("0123456789abcdef", 8*1024)
	p := newParser(bufio.NewReader(strings.NewReader(dataLength + "\x00\x00\x00\x00\x00" + data)))
	err := p.readData()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if !reflect.DeepEqual([]byte(data), p.Data) {
		t.Errorf("the data should be a * 128K, but %s", string(p.Data))
	}
}

func TestParser_readSymbol(t *testing.T) {
	in := "\x01\x02"
	p := newParser(bufio.NewReader(strings.NewReader(in)))
	err := p.readSymbol()
	if err != nil {
		t.Errorf("error should be nil")
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
		actual := reader.readVarint()
		if actual != testData.expected {
			t.Errorf("[%d] the value should be %d, but %d", i, testData.expected, actual)
		}
		if reader.err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, reader.err)
		}
		if reader.numReadBytes != int64(len(testData.in)) {
			t.Errorf("[%d] the number of read bytes should be %d, but %d", i, len(testData.in), reader.numReadBytes)
		}
	}
}

func TestReaderWithCounter_readVarint_Error(t *testing.T) {
	reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader(""))}
	_ = reader.readVarint()
	if reader.err == nil {
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
		actual := reader.readString()
		if actual != testData.expected {
			t.Errorf("[%d] the value should be %d, but %d", i, testData.expected, actual)
		}
		if reader.err != nil {
			t.Errorf("[%d] error should be nil, but %v", i, reader.err)
		}
		if reader.numReadBytes != int64(len(testData.expected)+1) {
			t.Errorf("[%d] the number of read bytes should be %d, but %d", i, len(testData.in), reader.numReadBytes)
		}
	}
}

func TestReaderWithCounter_readString_TooShortString(t *testing.T) {
	reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader("\x02"))}
	_ = reader.readString()
	if reader.err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestReaderWithCounter_readByte(t *testing.T) {
	value := "a"
	reader := readerWithCounter{raw: bufio.NewReader(strings.NewReader(value))}
	actual := reader.readByte()
	if actual != value[0] {
		t.Errorf("the value should be %v, but %v", value[0], actual)
	}
	if reader.err != nil {
		t.Errorf("error should be nil, but %v", reader.err)
	}
	if reader.numReadBytes != 1 {
		t.Errorf("the number of read bytes should be 1, but %d", reader.numReadBytes)
	}
}

func TestReaderWithCounter_read(t *testing.T) {
	value := []byte("abcdef")
	reader := readerWithCounter{raw: bufio.NewReader(bytes.NewReader(value))}
	buff := make([]byte, len(value))
	_ = reader.read(buff)
	if !reflect.DeepEqual(value, buff) {
		t.Errorf("the value should be %v, but %v", value, buff)
	}
	if reader.err != nil {
		t.Errorf("error should be nil, but %v", reader.err)
	}
	if reader.numReadBytes != int64(len(value)) {
		t.Errorf("the number of read bytes should be %d, but %d", len(value), reader.numReadBytes)
	}
}
