// +build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testDataDir = "testdata"

var cmdPath = filepath.Join(".", "readgoobj")

func TestMain(m *testing.M) {
	if err := exec.Command("go", "build").Run(); err != nil {
		fmt.Printf("failed to build tool: %+v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

var programList = []struct {
	name, expect string
}{
	{
		name: "helloworld.go",
		expect: `The list of defined symbols:
 Offset Size Type        DupOK Local MakeTypeLink Name                                       Version GoType
 0x0    0x6e STEXT       false false false        "".main                                    0
 0x8d   0x5b STEXT       false false false        "".init                                    0
 0x103  0x11 SRODATA     true  true  false        go.string."Hello, playground"              0
 0x114  0x21 SDWARFINFO  false false false        go.info."".main                            0
 0x135  0x0  SDWARFRANGE false false false        go.range."".main                           0
 0x135  0x21 SDWARFINFO  false false false        go.info."".init                            0
 0x156  0x0  SDWARFRANGE false false false        go.range."".init                           0
 0x156  0x10 SRODATA     false false false        "".statictmp_0                             0       type.string
 0x166  0x1  SNOPTRBSS   false false false        "".initdone·                              0       type.uint8
 0x166  0x1  SRODATA     true  true  false        runtime.gcbits.01                          0
 0x167  0x10 SRODATA     true  false false        type..namedata.*interface {}-              0
 0x177  0x38 SRODATA     true  false true         type.*interface {}                         0
 0x1af  0x1  SRODATA     true  true  false        runtime.gcbits.03                          0
 0x1b0  0x50 SRODATA     true  false false        type.interface {}                          0
 0x200  0x12 SRODATA     true  false false        type..namedata.*[]interface {}-            0
 0x212  0x38 SRODATA     true  false true         type.*[]interface {}                       0
 0x24a  0x38 SRODATA     true  false true         type.[]interface {}                        0
 0x282  0x13 SRODATA     true  false false        type..namedata.*[1]interface {}-           0
 0x295  0x38 SRODATA     true  false true         type.*[1]interface {}                      0
 0x2cd  0x48 SRODATA     true  false true         type.[1]interface {}                       0
 0x315  0x6  SRODATA     true  false false        type..importpath.fmt.                      0
 0x31b  0x8  SRODATA     true  false false        gclocals·69c1753bd5f81501d95132d08af04464 0
 0x323  0xa  SRODATA     true  false false        gclocals·e226d4ae4a7cad8835311c6a4683c14f 0
 0x32d  0x8  SRODATA     true  false false        gclocals·33cdeccccebe80329f1fdbee7f5874cb 0`},
}

func TestSamplePrograms(t *testing.T) {
	for i, program := range programList {
		goProgramName := filepath.Join(testDataDir, program.name)
		objectFileName := filepath.Join(testDataDir, strings.Replace(program.name, ".go", ".o", 1))
		if err := exec.Command("go", "tool", "compile", "-o", objectFileName, goProgramName).Run(); err != nil {
			t.Fatalf("[%d] failed to compile go file: %+v", i, err)
		}

		out, err := exec.Command(cmdPath, objectFileName).CombinedOutput()
		if err != nil {
			t.Fatalf("[%d] failed to run program\nerr: %v\nout: %v", i, err, string(out))
		}
		actualOutput := strings.Split(string(out), "\n")

		for j, expectedLine := range strings.Split(program.expect, "\n") {
			// ignore the number of spaces at the end of the line as it's trivial difference
			if strings.TrimRight(expectedLine, " ") != strings.TrimRight(actualOutput[j], " ") {
				t.Errorf("[%d] invalid output:\nexpect: %s\nactual: %s", i, expectedLine, actualOutput[j])
			}
		}
	}
}
