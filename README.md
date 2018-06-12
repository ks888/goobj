# goobj

[![CircleCI](https://circleci.com/gh/ks888/goobj.svg?style=svg)](https://circleci.com/gh/ks888/goobj)
[![Go Report Card](https://goreportcard.com/badge/github.com/ks888/goobj)](https://goreportcard.com/report/github.com/ks888/goobj)

A simple utility to read the contents of go object file.

## Prerequisite

Go 1.9 or 1.10

*Note: go object file is not formalized. This tool may not work well if the file format is updated in the future go releases.*

## Install

```
go get github.com/ks888/goobj
```

## Usage

Here is the simple go program.

```
% cat helloworld.go
package main

import (
   "fmt"
)

func main() {
   fmt.Println("Hello, playground")
}

```

Compile it. It will generate helloworld.o file.

```
% go tool compile helloworld.go
```

Check all the defined symbols in the object file. Note that the command name is *readgoobj* rather than *goobj*.

```
% readgoobj helloworld.o
The list of defined symbols:
 Offset Size Type        DupOK Local MakeTypeLink Name                                       Version GoType
 0x0    0x78 STEXT       false false false        "".main                                    0
 0x97   0x5b STEXT       false false false        "".init                                    0
 0x10d  0x11 SRODATA     true  true  false        go.string."Hello, playground"              0
 0x11e  0x1d SDWARFINFO  false false false        go.info."".main                            0
 0x13b  0x0  SDWARFRANGE false false false        go.range."".main                           0
 0x13b  0x1d SDWARFINFO  false false false        go.info."".init                            0
 0x158  0x0  SDWARFRANGE false false false        go.range."".init                           0
 0x158  0x10 SRODATA     false false false        "".statictmp_0                             0       type.string
 0x168  0x1  SNOPTRBSS   false false false        "".initdone·                              0       type.uint8
 0x168  0x1  SRODATA     true  true  false        runtime.gcbits.01                          0
 0x169  0x10 SRODATA     true  false false        type..namedata.*interface {}-              0
 0x179  0x38 SRODATA     true  false true         type.*interface {}                         0
 0x1b1  0x1  SRODATA     true  true  false        runtime.gcbits.03                          0
 0x1b2  0x50 SRODATA     true  false false        type.interface {}                          0
 0x202  0x12 SRODATA     true  false false        type..namedata.*[]interface {}-            0
 0x214  0x38 SRODATA     true  false true         type.*[]interface {}                       0
 0x24c  0x38 SRODATA     true  false true         type.[]interface {}                        0
 0x284  0x13 SRODATA     true  false false        type..namedata.*[1]interface {}-           0
 0x297  0x38 SRODATA     true  false true         type.*[1]interface {}                      0
 0x2cf  0x48 SRODATA     true  false true         type.[1]interface {}                       0
 0x317  0x6  SRODATA     true  false false        type..importpath.fmt.                      0
 0x31d  0x8  SRODATA     true  false false        gclocals·69c1753bd5f81501d95132d08af04464 0
 0x325  0xa  SRODATA     true  false false        gclocals·e226d4ae4a7cad8835311c6a4683c14f 0
 0x32f  0x8  SRODATA     true  false false        gclocals·33cdeccccebe80329f1fdbee7f5874cb 0
```