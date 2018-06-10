package goobj

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type GoObjParser struct {
	reader                  *bufio.Reader
	refs                    []SymbolReference
	readBytes, currDataAddr int64
}

func NewGoObjParser(f *os.File) *GoObjParser {
	return &GoObjParser{bufio.NewReader(f), nil, 0, 0}
}

func (p *GoObjParser) Parse() error {
	if err := p.skipToEndOfMagicHeader(); err != nil {
		return err
	}

	if err := p.skipVersion(); err != nil {
		return err
	}

	if err := p.skipDependencies(); err != nil {
		return err
	}

	var err error
	p.refs, err = p.readReferences()
	if err != nil {
		return err
	}

	dataLength, err := p.readVarint()
	if err != nil {
		return err
	}
	p.readVarint() // reloc
	p.readVarint() // pcdata
	p.readVarint() // automatics
	p.readVarint() // funcdata
	p.readVarint() // files
	p.currDataAddr = p.readBytes
	p.reader.Discard(int(dataLength))

	symbols, err := p.readSymbols()
	if err != nil {
		return err
	}

	fmt.Println("The list of defined symbols:")
	headers := []string{"Offset", "Size", "Type", "DupOK", "Local", "MakeTypeLink", "Name", "Version", "GoType"}
	table := NewTable(headers)
	for _, symbol := range symbols {
		ref := p.refs[symbol.IDIndex]
		goType := p.refs[symbol.GoTypeIndex]

		row := []string{
			fmt.Sprintf("%#x", symbol.Data.Offset),
			fmt.Sprintf("%#x", symbol.Size),
			fmt.Sprintf("%s", symbol.Kind),
			fmt.Sprintf("%v", symbol.DupOK),
			fmt.Sprintf("%v", symbol.Local),
			fmt.Sprintf("%v", symbol.Typelink),
			fmt.Sprintf("%s", ref.Name),
			fmt.Sprintf("%d", ref.Version),
			fmt.Sprintf("%s", goType.Name),
		}
		table.AddRow(row...)
	}
	table.PrintText(2)

	fmt.Println()

	fmt.Println("The optional fields of STEXT-typed symbols:")
	headers = []string{"Name", "FuncData"}
	table = NewTable(headers)
	for _, symbol := range symbols {
		if symbol.Kind != STEXT || symbol.stextFields == nil {
			continue
		}
		ref := p.refs[symbol.IDIndex]

		funcData := []string{}
		for i, funcDataIndex := range symbol.stextFields.FuncDataIndex {
			var associatedSymbol Symbol
			for _, candidate := range symbols {
				if candidate.IDIndex == funcDataIndex {
					associatedSymbol = candidate
					break
				}
			}

			funcData = append(funcData, fmt.Sprintf("%d - %s (%#x - %d)", i, p.refs[funcDataIndex].Name, associatedSymbol.Data.Offset, associatedSymbol.Data.Size))
		}

		row := []string{
			fmt.Sprintf("%s", ref.Name),
			strings.Join(funcData, ", "),
		}
		table.AddRow(row...)
	}
	table.PrintText(2)

	// fmt.Println("The list of relocations:")
	// headers = []string{"Offset", "Size", "Type", "SymbolName+Add"}
	// table = NewTable(headers)
	// for _, symbol := range symbols {
	// 	symName := p.refs[symbol.IDIndex].Name
	// 	for _, reloc := range symbol.Relocations {
	// 		row := []string{
	// 			fmt.Sprintf("%#x", reloc.Offset),
	// 			fmt.Sprintf("%#x", reloc.Size),
	// 			fmt.Sprintf("%s", reloc.Type),
	// 			fmt.Sprintf("%s+%d", symName, reloc.Add),
	// 		}
	// 		table.AddRow(row...)
	// 	}
	// }
	// table.PrintText(2)

	if err := p.skipMagicFooter(); err != nil {
		return err
	}

	return nil
}

func (p *GoObjParser) skipToEndOfMagicHeader() error {
	buff := make([]byte, len(magicHeader))
	n, err := p.reader.Read(buff)
	if err != nil {
		return err
	}
	p.readBytes += int64(n)

	for !reflect.DeepEqual(buff, magicHeader) {
		newByte, err := p.reader.ReadByte()
		if err != nil {
			return err
		}
		p.readBytes++

		buff = append(buff[1:], newByte)
	}

	return nil
}

func (p *GoObjParser) skipVersion() error {
	version, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	p.readBytes++

	if version != 1 {
		return fmt.Errorf("unexpected version: %d", version)
	}
	return nil
}

func (p *GoObjParser) skipDependencies() error {
	for {
		newByte, err := p.reader.ReadByte()
		if err != nil {
			return err
		}
		p.readBytes++

		if newByte == 0 {
			return nil
		}
	}
}

func (p *GoObjParser) readReferences() ([]SymbolReference, error) {
	refs := []SymbolReference{SymbolReference{}}
	for {
		newByte, err := p.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		p.readBytes++

		if newByte == 0xff {
			return refs, nil
		}

		if newByte != 0xfe {
			return nil, fmt.Errorf("sanity check failed: %#x ", newByte)
		}

		ref, err := p.readReference()
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
}

func (p *GoObjParser) readReference() (SymbolReference, error) {
	symbolName, err := p.readString()
	if err != nil {
		return SymbolReference{}, err
	}

	symbolVersion, err := p.readVarint()
	if err != nil {
		return SymbolReference{}, err
	}

	return SymbolReference{symbolName, symbolVersion}, nil
}

func (p *GoObjParser) readString() (string, error) {
	n, err := p.readVarint()
	if err != nil {
		return "", err
	}

	buff := make([]byte, n)
	numRead := int64(0)
	for numRead != n {
		read, err := p.reader.Read(buff[numRead:])
		if err != nil {
			return "", err
		}

		numRead += int64(read)
	}
	p.readBytes += numRead

	return string(buff), nil
}

func (p *GoObjParser) readVarint() (int64, error) {
	v := uint64(0)
	shift := uint(0)
	for {
		newByte, err := p.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		p.readBytes++

		v += uint64(newByte&0x7f) << shift
		if (newByte>>7)&0x1 == 0 {
			break
		}
		shift += 7
	}

	return zigzagDecode(v), nil
}

func (p *GoObjParser) readSymbols() ([]Symbol, error) {
	symbols := []Symbol{}
	for {
		newByte, err := p.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		if newByte == 0xff {
			return symbols, nil
		}

		if newByte != 0xfe {
			return nil, fmt.Errorf("sanity check failed: %#x ", newByte)
		}

		symbol, err := p.readSymbol()
		if err != nil {
			return nil, err
		}

		symbols = append(symbols, symbol)
	}
}

func (p *GoObjParser) readSymbol() (Symbol, error) {
	symbol := Symbol{}

	byte, err := p.reader.ReadByte()
	if err != nil {
		return symbol, err
	}
	symbol.Kind = SymKind(byte)

	symbol.IDIndex, err = p.readVarint()
	if err != nil {
		return symbol, err
	}

	flags, err := p.readVarint()
	if err != nil {
		return symbol, err
	}
	if flags&0x1 != 0 {
		symbol.DupOK = true
	}
	if (flags>>1)&0x1 != 0 {
		symbol.Local = true
	}
	if (flags>>2)&0x1 != 0 {
		symbol.Typelink = true
	}

	symbol.Size, err = p.readVarint()
	if err != nil {
		return symbol, err
	}

	symbol.GoTypeIndex, err = p.readVarint()
	if err != nil {
		return symbol, err
	}

	dataSize, err := p.readVarint()
	if err != nil {
		return symbol, err
	}
	symbol.Data = DataAddr{Size: int64(dataSize), Offset: p.currDataAddr}
	p.currDataAddr += dataSize

	numRelocs, err := p.readVarint()
	if err != nil {
		return symbol, err
	}

	for i := int64(0); i < numRelocs; i++ {
		reloc := Relocation{}
		reloc.Offset, _ = p.readVarint()
		reloc.Size, _ = p.readVarint()
		relocType, _ := p.readVarint()
		reloc.Type = RelocType(relocType)
		reloc.Add, _ = p.readVarint()
		reloc.IDIndex, _ = p.readVarint()

		symbol.Relocations = append(symbol.Relocations, reloc)
	}

	if symbol.Kind != STEXT {
		return symbol, nil
	}

	stextFields := &StextFields{}

	stextFields.Args, _ = p.readVarint()
	stextFields.Frame, _ = p.readVarint()
	flags, _ = p.readVarint()
	if flags&0x1 != 0 {
		stextFields.Leaf = true
	}
	if (flags>>1)&0x1 != 0 {
		stextFields.CFunc = true
	}
	if (flags>>2)&0x1 != 0 {
		stextFields.TypeMethod = true
	}
	if (flags>>3)&0x1 != 0 {
		stextFields.SharedFunc = true
	}
	noSplit, _ := p.readVarint()
	stextFields.NoSplit = noSplit != 0

	numLocals, _ := p.readVarint()

	for i := int64(0); i < numLocals; i++ {
		local := Local{}
		local.AsymIndex, _ = p.readVarint()
		local.Offset, _ = p.readVarint()
		local.Type, _ = p.readVarint()
		local.GotypeIndex, _ = p.readVarint()

		stextFields.Local = append(stextFields.Local, local)
	}

	pcspSize, _ := p.readVarint()
	stextFields.PCSP = DataAddr{Size: pcspSize, Offset: p.currDataAddr}
	p.currDataAddr += pcspSize

	pcFileSize, _ := p.readVarint()
	stextFields.PCFile = DataAddr{Size: pcFileSize, Offset: p.currDataAddr}
	p.currDataAddr += pcFileSize

	pcLineSize, _ := p.readVarint()
	stextFields.PCLine = DataAddr{Size: pcLineSize, Offset: p.currDataAddr}
	p.currDataAddr += pcLineSize

	pcInlineSize, _ := p.readVarint()
	stextFields.PCInline = DataAddr{Size: pcInlineSize, Offset: p.currDataAddr}
	p.currDataAddr += pcInlineSize

	numPCData, _ := p.readVarint()
	for i := int64(0); i < numPCData; i++ {
		pcDataSize, _ := p.readVarint()
		addr := DataAddr{Size: pcDataSize, Offset: p.currDataAddr}
		p.currDataAddr += pcDataSize

		stextFields.PCData = append(stextFields.PCData, addr)
	}

	numFuncData, _ := p.readVarint()
	for i := int64(0); i < numFuncData; i++ {
		funcDataIndex, _ := p.readVarint()
		stextFields.FuncDataIndex = append(stextFields.FuncDataIndex, funcDataIndex)
	}

	for i := int64(0); i < numFuncData; i++ {
		funcDataSym, _ := p.readVarint()
		stextFields.FuncDataOffset = append(stextFields.FuncDataOffset, funcDataSym)
	}

	numFiles, _ := p.readVarint()
	for i := int64(0); i < numFiles; i++ {
		fileIndex, _ := p.readVarint()
		stextFields.FileIndex = append(stextFields.FileIndex, fileIndex)
	}

	numInlineTrees, _ := p.readVarint()
	for i := int64(0); i < numInlineTrees; i++ {
		_, _ = p.readVarint() // parent
		_, _ = p.readVarint() // file
		_, _ = p.readVarint() // line
		_, _ = p.readVarint() // func
	}

	symbol.stextFields = stextFields

	return symbol, nil
}

func (p *GoObjParser) skipMagicFooter() error {
	buff := make([]byte, len(magicFooter))
	if _, err := p.reader.Read(buff); err != nil {
		return err
	}

	for !reflect.DeepEqual(buff, magicFooter) {
		return fmt.Errorf("magic footer not found: %v", buff)
	}

	return nil
}
