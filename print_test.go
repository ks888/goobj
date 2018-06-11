package goobj

import (
	"bytes"
	"reflect"
	"testing"
)

func TestTable_calcMaxWidths(t *testing.T) {
	table := newTable([]string{"a", "ab", "abc"})
	table.addRow([]string{"ab", "ab", "ab"})

	expected := []int{2, 2, 3}
	actual := table.calcMaxWidths()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("max widths should be %+v, but %+v", expected, actual)
	}
}

func TestTable_writeRowTo(t *testing.T) {
	table := newTable([]string{})
	buff := &bytes.Buffer{}
	expected := " a  ab \n"

	table.writeRowTo(buff, []string{"a", "ab"}, []int{2, 2})
	actual := buff.String()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("max widths should be %+v, but %+v", expected, actual)
	}
}
