/*
PICL.go: The PICL parser & code generator
*/

package PICL

import (
	"bufio"
   "bytes"
	"fmt"
	"picl-go/PICS"
)

type Object *ObjDesc
type ObjDesc struct {
	name               []byte
	form, typ, ptyp, a int
	next               Object
}

// Instruction tables for decoder
var table0 = [...]string{
	"MOVWF", "CLRF ", "SUBWF", "DECF ",
	"IORWF", "ANDWF", "XORWF", "ADDWF",
	"MOVFW", "COMF ", "INCF ", "DECFSZ",
	"RRF  ", "RLF  ", "SWAPF", "INCFSZ",
}

var table1 = [...]string{
	"BCF  ", "BSF  ", "BTFSC", "BTFSS",
}

var table2 = [...]string{
	"CALL ", "GOTO ",
}

var table3 = [...]string{
	"MOVLW", "", "", "",
	"RETLW", "", "", "",
	"IORLW", "ANDLW", "XORLW", "",
	"SUBLW", "ADDLW",
}

var (
	Dump                   bool
	sym                    int
	IdList, IdList0, undef Object
	pc, dc                 int
	err                    bool
)

// Parse error
func Mark(n int) {
	fmt.Printf("Parse error, code: %d\n", n)
	err = true
}

func Module() {
	var name = make([]byte, 0, 16)

	if sym == PICS.Module {
		PICS.Get(&sym)
		if sym == PICS.Ident {
			name = append(name, PICS.Id...)
			PICS.Get(&sym)
		} else {
			Mark(10)
		}
		if sym == PICS.Semicolon {
			PICS.Get(&sym)
		} else {
			Mark(20)
		}
	}

	if sym == PICS.Begin {
		PICS.Get(&sym)
		//StatSeq()
	}

	if sym == PICS.End {
		PICS.Get(&sym)
		if !(bytes.Compare(PICS.Id, name) == 0) {
			Mark(22)
		}
	} else {
		Mark(18)
	}
}

func Compile(reader *bufio.Reader) {
	IdList = IdList0
	PICS.Init(reader)
	pc = 1
	dc = 12
	err = false
	PICS.Get(&sym)
	Module()
}

// Add a new value to the symbol table
func enter(id string, form int, typ int, a int) {
	obj := new(ObjDesc)
	obj.name = make([]byte, 0, 16)
	obj.name = append(obj.name, id...)
	obj.form = form
	obj.typ = typ
	obj.a = a
	obj.next = IdList
	IdList = obj
}

// Run once on startup
func init() {
	undef = new(ObjDesc)
	enter("T", 1, 2, 1)
	enter("S", 1, 2, 3)
	enter("A", 1, 2, 5)
	enter("B", 1, 2, 6)
	IdList0 = IdList
	//Statement = Statement0
}
