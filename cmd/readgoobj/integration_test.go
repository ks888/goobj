// +build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const testDataDir = "testdata"

func TestMain(m *testing.M) {
	if err := exec.Command("go", "build").Run(); err != nil {
		fmt.Printf("failed to build tool: %+v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// Use already compiled program as an input because the object file is a little different by underlying OS and CPU.
var programList = []struct {
	name, expected string
}{
	{
		name: filepath.Join(testDataDir, "helloworld.o"),
		expected: `The list of defined symbols:
 Offset Size Type        DupOK Local MakeTypeLink Name                                       Version GoType
 0x3db  0x6e STEXT       false false false        "".main                                    0
 0x468  0x5b STEXT       false false false        "".init                                    0
 0x4de  0x11 SRODATA     true  true  false        go.string."Hello, playground"              0
 0x4ef  0x21 SDWARFINFO  false false false        go.info."".main                            0
 0x510  0x0  SDWARFRANGE false false false        go.range."".main                           0
 0x510  0x21 SDWARFINFO  false false false        go.info."".init                            0
 0x531  0x0  SDWARFRANGE false false false        go.range."".init                           0
 0x531  0x10 SRODATA     false false false        "".statictmp_0                             0       type.string
 0x541  0x1  SNOPTRBSS   false false false        "".initdone·                              0       type.uint8
 0x541  0x1  SRODATA     true  true  false        runtime.gcbits.01                          0
 0x542  0x10 SRODATA     true  false false        type..namedata.*interface {}-              0
 0x552  0x38 SRODATA     true  false true         type.*interface {}                         0
 0x58a  0x1  SRODATA     true  true  false        runtime.gcbits.03                          0
 0x58b  0x50 SRODATA     true  false false        type.interface {}                          0
 0x5db  0x12 SRODATA     true  false false        type..namedata.*[]interface {}-            0
 0x5ed  0x38 SRODATA     true  false true         type.*[]interface {}                       0
 0x625  0x38 SRODATA     true  false true         type.[]interface {}                        0
 0x65d  0x13 SRODATA     true  false false        type..namedata.*[1]interface {}-           0
 0x670  0x38 SRODATA     true  false true         type.*[1]interface {}                      0
 0x6a8  0x48 SRODATA     true  false true         type.[1]interface {}                       0
 0x6f0  0x6  SRODATA     true  false false        type..importpath.fmt.                      0
 0x6f6  0x8  SRODATA     true  false false        gclocals·69c1753bd5f81501d95132d08af04464 0
 0x6fe  0xa  SRODATA     true  false false        gclocals·e226d4ae4a7cad8835311c6a4683c14f 0
 0x708  0x8  SRODATA     true  false false        gclocals·33cdeccccebe80329f1fdbee7f5874cb 0`},
}

func TestSamplePrograms(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	var cmdPath = filepath.Join(filepath.Dir(filename), "readgoobj")

	for i, program := range programList {
		out, err := exec.Command(cmdPath, program.name).CombinedOutput()
		if err != nil {
			t.Fatalf("[%d] failed to run program\nerr: %v\nout: %v", i, err, string(out))
		}
		actualOutput := strings.Split(string(out), "\n")

		for j, expectedLine := range strings.Split(program.expected, "\n") {
			// ignore the number of spaces at the end of the line as it's trivial difference
			if strings.TrimRight(expectedLine, " ") != strings.TrimRight(actualOutput[j], " ") {
				t.Errorf("[%d] invalid output:\nexpect: %s\nactual: %s", i, expectedLine, actualOutput[j])
			}
		}
	}
}
