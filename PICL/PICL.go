/*
PICL.go: The PICL parser & Code generator
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
	variable  = 1
	constant  = 2
	procedure = 3
)

type Object *ObjDesc
type ObjDesc struct {
	name               []byte
	form, typ, ptyp, a int
	next               Object
}

var (
	sym                    int
	IdList, IdList0, undef Object
	Pc, dc                 int
	Err                    bool
	errs                   int
	Code                   [1024]int
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
	Err = true
	errs += 1
}

// Look up a name in the symbol table
// Linear search
func this(id []byte) Object {
	var obj Object

	obj = IdList
	for obj != nil && !(bytes.Compare(id, obj.name) == 0) {
		obj = obj.next
	}
	if obj == nil {
		Mark(10)
		obj = undef
	}
	return obj
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
	Code[Pc] = op*0x100 + a
	Pc += 1
}

// Put down BTFSS, BTFSC, BSF or BCF
func emit1(op int, n int, a int) {
	Code[Pc] = ((op+4)*8+n)*0x80 + a
	Pc += 1
}

// Handle bit selector in set notation
func index(n *int) {
	*n = 0
	if sym == PICS.Period {
		PICS.Get(&sym)
		if sym == PICS.Number {
			*n = PICS.Val
			PICS.Get(&sym)
		} else {
			Mark(11)
		}
	}
}

// Arithmetic expression handling
func expression() {
	var x, y Object
	var op, xf, xt, xval, yt, yval int

	// Object or literal?
	if sym == PICS.Ident {
		x = this(PICS.Id)
		xf = x.form
		xt = x.typ
		xval = x.a
		PICS.Get(&sym)
	} else if sym == PICS.Number {
		xf = constant
		xval = PICS.Val
		xt = PICS.Typ
		PICS.Get(&sym)
	} else {
		Mark(10)
		xval = 0
	}
	// Is it a function procedure?
	if sym == PICS.Lparen {
		PICS.Get(&sym)
		if x.form != procedure {
			Mark(3)
		}
		if sym != PICS.Rparen {
			expression()
		}
		emit(0x20, x.a)
		if sym == PICS.Rparen {
			PICS.Get(&sym)
		} else {
			Mark(8)
		}
	} else if (sym >= PICS.Ast) && (sym <= PICS.Minus) {
		// dyadic expression
		op = sym
		PICS.Get(&sym)
		yval = 0
		if sym == PICS.Ident {
			y = this(PICS.Id)
			yt = y.typ
			PICS.Get(&sym)
			// Instruction selection
			if y.form == variable {
				emit(0x08, y.a)
			} else if y.form == constant {
				emit(0x30, y.a)
			} else {
				Mark(10)
			}
		} else if sym == PICS.Number {
			yval = PICS.Val
			yt = PICS.Typ
			emit(0x30, yval)
			PICS.Get(&sym)
		}
		// Type check
		if xt != yt {
			Mark(111)
		}
		// Instruction selection
		if xf == variable {
			if op == PICS.Plus {
				if xt == PICS.Int_t {
					emit(0x07, x.a)
				} else {
					emit(0x04, x.a)
				}
			} else if op == PICS.Minus {
				if xt == PICS.Int_t {
					emit(0x02, x.a)
				} else {
					emit(0x06, x.a)
				}
			} else if op == PICS.Ast {
				if xt == PICS.Int_t {
					Mark(11)
				} else {
					emit(0x05, x.a)
				}
			}
		} else if xf == constant {
			if op == PICS.Plus {
				if xt == PICS.Int_t {
					emit(0x3E, xval)
				} else {
					emit(0x38, xval)
				}
			} else if op == PICS.Minus {
				if xt == PICS.Int_t {
					emit(0x3C, xval)
				} else {
					emit(0x3A, xval)
				}
			} else if op == PICS.Ast {
				if xt == PICS.Int_t {
					Mark(11)
				} else {
					emit(0x39, xval)
				}
			} else {
				Mark(9)
			}
		} else {
			Mark(10)
		}
	} else if xf == variable {
		emit(0x08, x.a)
	} else if xf == constant {
		emit(0x30, xval)
	} else {
		Mark(10)
	}
}

// Logical expression handling
func term() {
	var x, y Object
	var n, rel, yf, ya int

	if sym == PICS.Ident {
		x = this(PICS.Id)
		PICS.Get(&sym)
		if (sym >= PICS.Eql) && (sym <= PICS.Gtr) {
			rel = sym
			PICS.Get(&sym)
			if sym == PICS.Ident {
				y = this(PICS.Id)
				PICS.Get(&sym)
				yf = y.form
				ya = y.a
			} else if sym == PICS.Number {
				yf = constant
				ya = PICS.Val
				PICS.Get(&sym)
			}
			if rel < PICS.Leq {
				if yf == variable {
					emit(0x08, ya)
					emit(0x02, x.a)
				} else if yf == constant {
					if ya == 0 {
						emit(0x08, x.a)
					} else {
						emit(0x30, ya)
						emit(0x02, x.a)
					}
				}
			} else {
				emit(0x08, x.a)
				if yf == variable {
					emit(0x02, ya)
				} else if (yf == constant) && (yf != 0) {
					emit(0x60, ya)
				}
			}
			if rel == PICS.Eql {
				emit1(3, 2, 3)
			} else if rel == PICS.Neq {
				emit1(2, 2, 3)
			} else if (rel == PICS.Geq) || (rel == PICS.Leq) {
				emit1(3, 0, 3)
			} else if (rel == PICS.Lss) || (rel == PICS.Gtr) {
				emit1(2, 0, 3)
			}
		} else {
			index(&n)
			emit1(3, n, x.a)
		}
	} else if sym == PICS.Not {
		PICS.Get(&sym)
		if sym == PICS.Ident {
			x = this(PICS.Id)
			PICS.Get(&sym)
			index(&n)
			emit1(2, n, x.a)
		} else {
			Mark(10)
		}
	} else {
		Mark(10)
	}
}

// Conditional expression for guarded statements
func condition(link *int) {
	var L, L0, L1 int

	term()
	Code[Pc] = 0
	L = Pc
	Pc += 1

	if sym == PICS.And {
		for {
			PICS.Get(&sym)
			term()
			Code[Pc] = L
			L = Pc
			Pc += 1
			if sym != PICS.And {
				break
			}
		}
	} else if sym == PICS.Or {
		for {
			PICS.Get(&sym)
			term()
			Code[Pc] = L
			L = Pc
			Pc += 1
			if sym != PICS.Or {
				break
			}
		}
		L0 = Code[L]
		Code[L] = 0
		for {
			if (Code[L0-1] / 0x400) == 6 {
				Code[L0-1] += 0x400
			} else {
				Code[L0-1] -= 0x400
			}
			L1 = Code[L0]
			Code[L0] = Pc + 0x2800
			L0 = L1
			if L0 == 0 {
				break
			}
		}
	}
	*link = L
}

// Fix up forward and backward jumps
func fixup(L int, k int) {
	var L1 int

	for L != 0 {
		L1 = Code[L]
		Code[L] = k + 0x2800
		L = L1
	}
}

// Statement sequence
// NOTE: Statement() is not a pointer indirection
func StatSeq() {
	Statement()
	for sym == PICS.Semicolon {
		PICS.Get(&sym)
		Statement()
	}
}

// Guarded statement block
// NOTE: not actually called anywhere in original code, but if condition terminating
// symbol is made a param, can be used for IF, ELSIF and WHILE blocks
func Guarded(s int, L *int) {
	condition(L)
	if sym == s {
		PICS.Get(&sym)
	} else {
		Mark(14)
	}
	StatSeq()
}

// Conditional Statements
func IfStat() {
	var L0, L int

	Guarded(PICS.Then, &L)
	L0 = 0
	for sym == PICS.Elsif {
		Code[Pc] = L0
		L0 = Pc
		Pc += 1
		fixup(L, Pc)
		PICS.Get(&sym)
		Guarded(PICS.Then, &L)
	}
	if sym == PICS.Else {
		Code[Pc] = L0
		L0 = Pc
		Pc += 1
		fixup(L, Pc)
		PICS.Get(&sym)
		StatSeq()
	} else {
		fixup(L, Pc)
	}
	if sym == PICS.End {
		PICS.Get(&sym)
	} else {
		Mark(15)
	}
	fixup(L0, Pc)
}

// Conditional Repetition: condition first
func WhileStat() {
	var L0, L int

	L0 = Pc
	Guarded(PICS.Do, &L)
	emit(0x28, L0)
	fixup(L, Pc)
	for sym == PICS.Elsif {
		PICS.Get(&sym)
		Guarded(PICS.Do, &L)
		emit(0x28, L0)
		fixup(L, Pc)
	}
	if sym == PICS.End {
		PICS.Get(&sym)
	} else {
		Mark(16)
	}
}

// Conditional Repetition: condition last
func RepeatStat() {
	var L0, L int

	L0 = Pc
	StatSeq()
	if sym == PICS.Until {
		PICS.Get(&sym)
		condition(&L)
		if (Code[Pc-4]/0x100 == 3) && (Code[Pc-3]/0x100 == 8) &&
			(Code[Pc-2] == 0x1D03) && (Code[Pc-4]%0x80 == Code[Pc-3]%0x100) {
			Code[Pc-4] += 0x800
			Code[Pc-3] = 0
			Pc -= 2
			L = Pc - 1
		}
		fixup(L, L0)
	} else if sym == PICS.End {
		PICS.Get(&sym)
		emit(0x28, L0)
	} else {
		Mark(25)
	}
}

// Assignment Statement (new)
// NOTE: factored out from Statement() vs original code
func AssignStat(x Object) {
	var w int

	PICS.Get(&sym)
	if x.form != variable {
		Mark(2)
	}
	expression()
	w = Code[Pc-1]
	if w == 0x3000 {
		Code[Pc-1] = x.a + 0x180
	} else if ((w / 0x100) <= 13) && (w%0x100 == x.a) {
		Code[Pc-1] += 0x80
	} else {
		emit(0, x.a+0x80)
	}
}

// Procedure Call Statement (new)
// NOTE: factored out from Statement() vs original code
func CallStat(x Object) {
	if x.form != procedure {
		Mark(3)
	}
	if sym == PICS.Lparen {
		PICS.Get(&sym)
		expression()
		emit(0x20, x.a)
		if sym == PICS.Rparen {
			PICS.Get(&sym)
		} else {
			Mark(8)
		}
	} else {
		emit(0x20, x.a)
	}
}

func Operand1(cd int) {
	var x Object

	if sym == PICS.Ident {
		x = this(PICS.Id)
		PICS.Get(&sym)
		if x.form != variable {
			Mark(2)
		}
		emit(cd, x.a+0x80)
	} else {
		Mark(10)
	}
}

func Operand2(cd int) {
	var x Object
	var n int

	if sym == PICS.Ident {
		x = this(PICS.Id)
		PICS.Get(&sym)
		if x.form != variable {
			Mark(2)
		}
		index(&n)
		emit1(cd, n, x.a)
	} else {
		Mark(10)
	}
}

// Statement
// NOTE: renamed from Statement0
func Statement() {
	var x Object

	switch sym {
	case PICS.Ident:
		x = this(PICS.Id)
		PICS.Get(&sym)
		if sym == PICS.Becomes {
			AssignStat(x)
		} else {
			CallStat(x)
		}
	case PICS.Inc:
		PICS.Get(&sym)
		Operand1(10)
	case PICS.Dec:
		PICS.Get(&sym)
		Operand1(3)
	case PICS.Rol:
		PICS.Get(&sym)
		Operand1(13)
	case PICS.Ror:
		PICS.Get(&sym)
		Operand1(12)
	case PICS.Op:
		PICS.Get(&sym)
		if sym == PICS.Not {
			PICS.Get(&sym)
			Operand2(0)
		} else {
			Operand2(1)
		}
	case PICS.Query:
		PICS.Get(&sym)
		if sym == PICS.Not {
			PICS.Get(&sym)
			Operand2(2)
		} else {
			Operand2(3)
		}
		emit(0x28, Pc-1)
	case PICS.Lparen:
		PICS.Get(&sym)
		StatSeq()
		if sym == PICS.Rparen {
			PICS.Get(&sym)
		} else {
			Mark(8)
		}
	case PICS.If:
		PICS.Get(&sym)
		IfStat()
	case PICS.While:
		PICS.Get(&sym)
		WhileStat()
	case PICS.Repeat:
		PICS.Get(&sym)
		RepeatStat()
	}
}

// Procedure declarations
func ProcDecl() {
	var typ, partyp, restyp, pc0 int
	var obj Object
	var name = make([]byte, 0, 16)

	obj = IdList
	partyp = 0
	restyp = 0
	pc0 = Pc

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
		if (sym >= PICS.Int) && (sym <= PICS.Bool) {
			partyp = sym - PICS.Int + 1
			PICS.Get(&sym)
			if sym == PICS.Ident {
				enter(string(PICS.Id), variable, partyp, dc)
				PICS.Get(&sym)
				emit(0, dc+0x80)
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
		if (sym >= PICS.Int) && (sym <= PICS.Bool) {
			restyp = sym - PICS.Int + 1
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
	for (sym >= PICS.Int) && (sym <= PICS.Bool) {
		typ = sym - PICS.Int + 1
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
		StatSeq()
	} else {
		Mark(21)
	}
	if sym == PICS.Return {
		PICS.Get(&sym)
		expression()
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
	if sym == PICS.Const {
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
	for (sym >= PICS.Int) && (sym <= PICS.Bool) {
		typ = sym - PICS.Int + 1
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

	if Pc > 1 {
		Code[0] = Pc + 0x2800
	} else {
		Pc = 0
	}

	// Module body
	if sym == PICS.Begin {
		PICS.Get(&sym)
		StatSeq()
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


// Entry point for module
func Compile(reader *bufio.Reader) {
	IdList = IdList0
	PICS.Init(reader)
	Pc = 1
	dc = 12
	Err = false
	PICS.Get(&sym)
	Module()
	fmt.Printf("Errors: %d\n", errs)
}

// Run once on startup
func init() {
	Err = false
	errs = 0
	undef = new(ObjDesc)
	enter("T", variable, PICS.Set_t, 0x01)
	enter("S", variable, PICS.Set_t, 0x03)
	enter("A", variable, PICS.Set_t, 0x05)
	enter("B", variable, PICS.Set_t, 0x06)
	IdList0 = IdList
}
