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

// Item forms
const (
   variable = 1
   constant = 2
   procedure = 3
)
   
type Object *ObjDesc
type ObjDesc struct {
	name               []byte
	form, typ, ptyp, a int
	next               Object
}

var (
	Dump                   bool
	sym                    int
	IdList, IdList0, undef Object
	pc, dc                 int
	err                    bool
   code                   [1024]int
)

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

// Parse error
func Mark(n int) {
	fmt.Printf("Parse error, code: %d\n", n)
	err = true
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

// Put down a regular opcode
func emit(op int, a int) {
   code[pc] = op * 0x100 + a
   pc += 1
}

// Put down BTFSS, BTFSC, BSF or BCF
func emit1(op int, n int, a int) {
   code[pc] = ((op + 4) * 8 + n) * 0x80 + a
   pc += 1
}
   
// Procedure declarations
func ProcDecl() {
   var typ, partyp, restyp, pc0 int
   var obj Object
   var name = make([]byte, 0, 16)
   
   obj = IdList
   partyp = 0
   restyp = 0
   pc0 = pc
   
   // Procedure name
   if sym == PICS.Ident {
      name = append(name, PICS.Id...)
      PICS.Get(&sym)
   } else {
      Mark(10)
   }
   
   // Optional parens with optional argument
   if sym == PICS.Lparen {
      PICS.Get(&sym)
      if (sym >= PICS.Int_) && (sym <= PICS.Bool_) {
         partyp = sym - PICS.Int_ + 1
         PICS.Get(&sym)
         if sym == PICS.Ident {
            enter(string(PICS.Id), variable, partyp, dc)
            PICS.Get(&sym)
            emit(0, dc + 0x80)
            dc += 1
         } else {
            Mark(10)
         }
      }
      if sym == PICS.Rparen {
         PICS.Get(&sym)
      } else {
         Mark(8)
      }
   }
   
   // Optional result type
   if sym == PICS.Colon {
      PICS.Get(&sym)
      if (sym >= PICS.Int_) && (sym <= PICS.Bool_) {
         restyp = sym - PICS.Int_ + 1
         PICS.Get(&sym)
      } else {
         Mark(10)
      }
   }
   
   // Terminate procedure header
   if sym == PICS.Semicolon {
      PICS.Get(&sym)
   } else {
      Mark(20)
   }
   
   // Variable declarations
   for (sym >= PICS.Int_) && (sym <= PICS.Bool_) {
      typ = sym - PICS.Int_ + 1
      PICS.Get(&sym)
      for sym == PICS.Ident {
         enter(string(PICS.Id), variable, typ, dc)
         dc += 1
         PICS.Get(&sym)
         if sym == PICS.Comma {
            PICS.Get(&sym)
         }
      }
      if sym == PICS.Semicolon {
         PICS.Get(&sym)
      } else {
         Mark(20)
      }
   }
   
   // Procedure body
   if sym == PICS.Begin {
      PICS.Get(&sym)
      //StatSeq()
   } else {
      Mark(21)
   }
   if sym == PICS.Return_ {
      PICS.Get(&sym)
      //expression()
   }
   emit(0, 8)
   if sym == PICS.End {
      PICS.Get(&sym)
      if sym == PICS.Ident {
         if !(bytes.Compare(PICS.Id, name) == 0) {
            Mark(22)
         }
         PICS.Get(&sym)
      } else {
         Mark(10)
      }
   } else {
      Mark(18)
   }
   if sym == PICS.Semicolon {
      PICS.Get(&sym)
   } else {
      Mark(20)
   }
   
   // Clean up
   IdList = obj
   enter(string(name), procedure, restyp, pc0)
   IdList.ptyp = partyp
}
      
func Module() {
   var typ int
	var name = make([]byte, 0, 16)

   // Module header
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
   
   // CONST Declarations
   if sym == PICS.Const_ {
      PICS.Get(&sym)
      for sym == PICS.Ident {
         enter(string(PICS.Id), constant, 1, 0)
         PICS.Get(&sym)
         if sym == PICS.Eql {
            PICS.Get(&sym)
            if sym == PICS.Number {
               IdList.a = PICS.Val
               PICS.Get(&sym)
            } else {
               Mark(10)
            }
         } else {
            Mark(5)
         }
         if sym == PICS.Semicolon {
            PICS.Get(&sym)
         } else {
            Mark(20)
         }
      }
   }
   
   // Var Declarations: INT, BOOL, SET
   for (sym >= PICS.Int_) && (sym <= PICS.Bool_) {
      typ = sym - PICS.Int_ + 1
      PICS.Get(&sym)
      // May be a list of identifiers eg INT a, b, c
      for sym == PICS.Ident {
         enter(string(PICS.Id), variable, typ, dc)
         dc += 1
         PICS.Get(&sym)
         if sym == PICS.Comma {
            PICS.Get(&sym)
         }
      }
      // Optional semicolon after var declaration?
      if sym == PICS.Semicolon {
         PICS.Get(&sym)
      }
   }
   
   // PROCEDURE Declarations
   for sym == PICS.Proced {
      PICS.Get(&sym)
      ProcDecl()
   }
      
   if pc > 1 {
      code[0] = pc + 0x2800
   } else {
      pc = 0
   }
   
   // Module body
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
