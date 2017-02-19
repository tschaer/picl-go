/*
Driver file for N. Wirth's PICL compiler

Build with:
C:\> go install picl/piclc
Run:
C:\Data\Personal\go\bin> piclc file.asm
*/

package main

import (
	"picl-go/PICL"
	"bufio"
	"flag"
	"fmt"
	"os"
)

const Ver = "PICL compiler v0.1a"

func init() {
	flag.BoolVar(&PICL.Dump, "d", false, "Dump listing output to console")
}

func main() {

	fmt.Printf("%s\n\n", Ver)

	// Handle args
	flag.Parse()
	// Exit on error
	if !(len(flag.Args()) > 0) {
		fmt.Printf("Usage: piclc <flags> sourcefile.asm\n")
		flag.PrintDefaults()
		return
	}

	// Compile source file
	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Gonna compile %s\n", filename)
		PICL.Compile(bufio.NewReader(file))
	}
}
