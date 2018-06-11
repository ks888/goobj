package goobj

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var symbolHeaderRows = []string{"Offset", "Size", "Type", "DupOK", "Local", "MakeTypeLink", "Name", "Version", "GoType"}

// PrintSymbols prints the symbols in the table format.
func PrintSymbols(file *File) {
	fmt.Println("The list of defined symbols:")

	table := newTable(symbolHeaderRows)
	for _, symbol := range file.symbols {
		ref := file.symbolReferences[symbol.IDIndex]
		goType := file.symbolReferences[symbol.GoTypeIndex]

		row := []string{
			fmt.Sprintf("%#x", symbol.DataAddr.Offset),
			fmt.Sprintf("%#x", symbol.Size),
			fmt.Sprintf("%s", symbol.Kind),
			fmt.Sprintf("%v", symbol.DupOK),
			fmt.Sprintf("%v", symbol.Local),
			fmt.Sprintf("%v", symbol.Typelink),
			fmt.Sprintf("%s", ref.Name),
			fmt.Sprintf("%d", ref.Version),
			fmt.Sprintf("%s", goType.Name),
		}
		table.addRow(row)
	}
	table.print()
}

type table struct {
	headers []string
	rows    [][]string
}

func newTable(headers []string) *table {
	return &table{headers: headers}
}

func (t *table) addRow(values []string) {
	t.rows = append(t.rows, values)
}

func (t *table) print() {
	t.writeTo(os.Stdout)
}

func (t *table) writeTo(w io.Writer) {
	maxWidths := t.calcMaxWidths()

	t.writeRowTo(w, t.headers, maxWidths)
	for _, row := range t.rows {
		t.writeRowTo(w, row, maxWidths)
	}
}

func (t *table) calcMaxWidths() []int {
	maxWidths := make([]int, len(t.headers))
	for i, header := range t.headers {
		maxWidths[i] = len(header)
	}

	for _, row := range t.rows {
		for j, val := range row {
			if maxWidths[j] < len(val) {
				maxWidths[j] = len(val)
			}
		}
	}

	return maxWidths
}

func (t *table) writeRowTo(w io.Writer, row []string, maxWidths []int) {
	fmt.Fprint(w, " ")
	for i, val := range row {
		fmt.Fprint(w, val)
		fmt.Fprint(w, strings.Repeat(" ", maxWidths[i]-len(val)+1))
	}
	fmt.Fprint(w, "\n")
}
