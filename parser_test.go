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

func TestParser_parseReferences(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\xfe\x02a\x02\xff")))
	err := p.parseReferences()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.SymbolReferences) != 2 {
		t.Errorf("the number of symbolReferences should be 2, but %d", len(p.SymbolReferences))
	}
	expect := SymbolReference{Name: "", Version: 0}
	if p.SymbolReferences[0] != expect {
		t.Errorf("invalid symbolReference: %+v", p.SymbolReferences[0])
	}
	if p.SymbolReferences[1].Name != "a" || p.SymbolReferences[1].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.SymbolReferences[1])
	}
}

func TestParser_parseReference(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x02a\x02")))
	err := p.parseReference()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.SymbolReferences) != 1 {
		t.Errorf("the number of symbolReferences should be 1, but %d", len(p.SymbolReferences))
	}
	if p.SymbolReferences[0].Name != "a" || p.SymbolReferences[0].Version != 1 {
		t.Errorf("invalid symbolReference: %+v", p.SymbolReferences[0])
	}
}

func TestParser_parseData(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\x02\x00\x00\x00\x00\x00a")))
	err := p.parseData()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if !reflect.DeepEqual([]byte("a"), p.Data) {
		t.Errorf("the data should be a, but %s", string(p.Data))
	}
}

func TestParser_parseData_128KBData(t *testing.T) {
	dataLength := "\x80\x80\x10" // 128KB
	data := strings.Repeat("0123456789abcdef", 8*1024)
	p := newParser(bufio.NewReader(strings.NewReader(dataLength + "\x00\x00\x00\x00\x00" + data)))
	err := p.parseData()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if !reflect.DeepEqual([]byte(data), p.Data) {
		t.Errorf("the data should be a * 128K, but %s", string(p.Data))
	}
}

func TestParser_parseSymbols(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("\xfe" + symbolForTesting() + "\xff")))
	err := p.parseSymbols()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.Symbols) != 1 {
		t.Errorf("the number of symbols should be 1, but %d", len(p.Symbols))
	}
}

func symbolForTesting() string {
	return "\x02\x02\x0e\x02\x02\x02\x02\x02\x02\x02\x02\x02"
}

func TestParser_parseSymbol(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader(symbolForTesting())))
	err := p.parseSymbol()
	if err != nil {
		t.Errorf("error should be nil")
	}
	if len(p.Symbols) != 1 {
		t.Errorf("the number of symbols should be 1, but %d", len(p.Symbols))
	}

	actual := p.Symbols[0]
	if SRODATA != actual.Kind {
		t.Errorf("the kind should be %s, but %s", STEXT, actual.Kind)
	}
	if actual.IDIndex != 1 {
		t.Errorf("the id index should be 1, but %d", actual.IDIndex)
	}
	if !actual.DupOK {
		t.Errorf("DupOK flag should be true")
	}
	if !actual.Local {
		t.Errorf("Local flag should be true")
	}
	if !actual.Typelink {
		t.Errorf("Typelink flag should be true")
	}
	if actual.Size != 1 {
		t.Errorf("the size should be 1, but %d", actual.Size)
	}
	if actual.GoTypeIndex != 1 {
		t.Errorf("the GoType index should be 1, but %d", actual.GoTypeIndex)
	}
	expectedData := DataAddr{Size: 1, Offset: 0}
	if expectedData != actual.DataAddr {
		t.Errorf("the data should be %+v, but %+v", expectedData, actual.DataAddr)
	}
	if p.associatedDataSize != 1 {
		t.Errorf("the associatedDataSize should be 1, but %d", p.associatedDataSize)
	}
	if len(actual.Relocations) != 1 {
		t.Errorf("the number of relocations should be 1, but %d", len(actual.Relocations))
	}
	expectedReloc := Relocation{Offset: 1, Size: 1, Type: 1, Add: 1, IDIndex: 1}
	if actual.Relocations[0] != expectedReloc {
		t.Errorf("the relocation should be %+v, but %+v", expectedReloc, actual.Relocations[0])
	}
}

func TestParser_parseSymbol_EmptyInput(t *testing.T) {
	p := newParser(bufio.NewReader(strings.NewReader("")))
	err := p.parseSymbol()
	if err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestParser_skipSTEXTFields(t *testing.T) {
	in := "\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x02\x00\x00\x02\x00\x02\x00\x00\x00\x00"
	p := newParser(bufio.NewReader(strings.NewReader(in)))
	err := p.skipSTEXTFields()
	if err != nil {
		t.Errorf("error should be nil")
	}
}

func TestParser_skipFooter(t *testing.T) {
	p := newParser(bufio.NewReader(bytes.NewReader(magicFooter)))
	err := p.skipFooter()
	if err != nil {
		t.Errorf("error should be nil")
	}
}

func TestParser_skipFooter_wrongFooter(t *testing.T) {
	p := newParser(bufio.NewReader(bytes.NewReader(append([]byte("\xfe"), magicFooter...))))
	err := p.skipFooter()
	if err == nil {
		t.Errorf("error should not be nil")
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
			t.Errorf("[%d] the value should be %s, but %s", i, testData.expected, actual)
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
