package goobj

import (
	"fmt"
	"strings"
)

type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers []string) *Table {
	return &Table{headers: headers}
}

func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

func (t *Table) PrintText(numIndentSpaces int) error {
	maxWidths := t.calcMaxWidths()

	t.printRow(t.headers, maxWidths, numIndentSpaces)
	for _, row := range t.rows {
		t.printRow(row, maxWidths, numIndentSpaces)
	}

	return nil
}

func (t *Table) calcMaxWidths() []int {
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

func (t *Table) printRow(row []string, maxWidths []int, numIndentSpaces int) {
	fmt.Print(strings.Repeat(" ", numIndentSpaces))
	for i, val := range row {
		fmt.Print(val)
		fmt.Print(strings.Repeat(" ", maxWidths[i]-len(val)+1))
	}
	fmt.Print("\n")
}
