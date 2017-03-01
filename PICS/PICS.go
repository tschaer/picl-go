/*
PICS.go: The PICL Scanner
Notes:
1. Texts.Read(R, ch) -> ch, err = r.ReadByte()
2. No error checking!! This from the original source
3. Go's init syntax used to set key & synmo, ditched Enter()
4. Symbol constants are exported instead of duplicating in PICL
5. Numeric type constants duplicated here
*/

package PICS

import (
	"bufio"
	//"fmt"
	"io"
)

const IdLen = 32

// Numeric types
const (
	Int_t  = 1
	Set_t  = 2
	Bool_t = 3
)

// Symbols
const (
	Null      = 0
	Ast       = 1
	Slash     = 2
	Plus      = 3
	Minus     = 4
	Not       = 5
	And       = 6
	Or        = 7
	Eql       = 10
	Neq       = 11
	Geq       = 12
	Lss       = 13
	Leq       = 14
	Gtr       = 15
	Period    = 16
	Comma     = 17
	Colon     = 18
	Op        = 20
	Query     = 21
	Lparen    = 22
	Becomes   = 23
	Ident     = 24
	If        = 25
	While     = 26
	Repeat    = 27
	Inc       = 28
	Dec       = 29
	Rol       = 30
	Ror       = 31
	Number    = 32
	Rparen    = 33
	Then      = 34
	Do        = 35
	Semicolon = 36
	End       = 37
	Else      = 38
	Elsif     = 39
	Until     = 40
	Return    = 41
	Int       = 42
	Set       = 43
	Bool      = 44
	Const     = 50
	Begin     = 51
	Proced    = 52
	Module    = 53
	Eof       = 54
)

var (
	ch  byte
	err error
	r   *bufio.Reader
	Val int
	Typ int
	Id  []byte
)

// key & symno are the table of recognised symbols in the PICL grammar
// NOTE!! must be sorted, binary search is used
var key = [...]string{
	"BEGIN", "BOOL", "CONST", "DEC",
	"DO", "ELSE", "ELSIF", "END",
	"IF", "INC", "INT", "MODULE",
	"OR", "PROCEDURE", "REPEAT", "RETURN",
	"ROL", "ROR", "SET", "THEN",
	"UNTIL", "WHILE", "~ ",
}
var symno = [...]int{
	Begin, Bool, Const, Dec,
	Do, Else, Elsif, End,
	If, Inc, Int, Module,
	Or, Proced, Repeat, Return,
	Rol, Ror, Set, Then,
	Until, While,
}

// Handle identifiers and keywords
func identifier() int {
	// Zero out last usage
	Id = Id[:0]
	i := 0

	// Read in contiguous alphanum chars
	for {
		if i < 16 {
			Id = append(Id, ch)
			i += 1
		}
		ch, _ = r.ReadByte()
		if (ch < '0') ||
			(ch > '9' && ch < 'A') ||
			(ch > 'Z' && ch < 'a') ||
			(ch > 'z') {
			break
		}
	}

	// Search keyword table
	i = 0
	j := len(key)
	for i < j {
		// binary search
		m := (i + j) / 2
		if key[m] < string(Id[:]) {
			i = m + 1
		} else {
			j = m
		}
	}

	// Identifier or keyword?
	if key[j] == string(Id[:]) {
		return symno[i]
	}
	return Ident
}

// Get a decimal number
func number() {
	Val = 0
	for {
		Val = 10*Val + int(ch-'0')
		ch, err = r.ReadByte()
		if (ch < '0') || (ch > '9') {
			break
		}
	}
}

// Helper for hex()
func getDigit() int {
	var d int

	if (ch >= '0') && (ch <= '9') {
		d = int(ch - '0')
	} else if ch >= 'A' && ch <= 'F' {
		d = int(ch - '7')
	} else {
		d = 0
	}
	ch, err = r.ReadByte()

	return d
}

// Get a SET literal ($xx)
func hex() {
	Val = getDigit()<<4 | getDigit()
}

// Return next symbol in input
func Get(sym *int) {
	// Eat whitespace or anything enclosed in {}
	for (ch <= ' ') || (ch == '{') {
		if ch == '{' {
			for {
				ch, err = r.ReadByte()
				if (ch == '}') || (err == io.EOF) {
					break
				}
			}
		}
		ch, _ = r.ReadByte()
	}
	// Repeat until a valid symbol is found (this includes EOF)
	for {
		// Eat whitespace
		for err != io.EOF && (ch <= ' ') {
			ch, err = r.ReadByte()
		}
		// Recognise symbol
		if err == io.EOF {
			*sym = Eof
		} else {
			switch {
			case ch == '!':
				ch, err = r.ReadByte()
				*sym = Op
			case ch == '#':
				ch, err = r.ReadByte()
				*sym = Neq
			case ch == '$':
				ch, err = r.ReadByte()
				hex()
				*sym = Number
				Typ = Set_t
			case ch == '&':
				ch, err = r.ReadByte()
				*sym = And
			case ch == '(':
				ch, err = r.ReadByte()
				*sym = Lparen
			case ch == ')':
				ch, err = r.ReadByte()
				*sym = Rparen
			case ch == '*':
				ch, err = r.ReadByte()
				*sym = Ast
			case ch == '+':
				ch, err = r.ReadByte()
				*sym = Plus
			case ch == ',':
				ch, err = r.ReadByte()
				*sym = Comma
			case ch == '-':
				ch, err = r.ReadByte()
				*sym = Minus
			case ch == '.':
				ch, err = r.ReadByte()
				*sym = Period
			case ch == '/':
				ch, err = r.ReadByte()
				*sym = Slash
			case ch >= '0' && ch <= '9':
				number()
				*sym = Number
				Typ = Int_t
			case ch == ':':
				ch, err = r.ReadByte()
				if ch == '=' {
					ch, err = r.ReadByte()
					*sym = Becomes
				} else {
					*sym = Colon
				}
			case ch == ';':
				ch, err = r.ReadByte()
				*sym = Semicolon
			case ch == '<':
				ch, err = r.ReadByte()
				if ch == '=' {
					ch, err = r.ReadByte()
					*sym = Leq
				} else {
					*sym = Lss
				}
			case ch == '=':
				ch, err = r.ReadByte()
				*sym = Eql
			case ch == '>':
				ch, err = r.ReadByte()
				if ch == '=' {
					ch, err = r.ReadByte()
					*sym = Geq
				} else {
					*sym = Gtr
				}
			case ch == '?':
				ch, err = r.ReadByte()
				*sym = Query
			case ch == '~':
				ch, err = r.ReadByte()
				*sym = Not
			case (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z'):
				*sym = identifier()
			default:
				ch, err = r.ReadByte()
				*sym = Null
			}
		}
		// Exit if symbol is valid, otherwise try again
		if *sym != Null {
			break
		}
	}
	//fmt.Printf("Gets(): sym = %d\n", *sym)
}

// Scanner init
func Init(reader *bufio.Reader) {
	r = reader
	ch, _ = r.ReadByte()
}

// Run once at startup
func init() {
	Id = make([]byte, 16)
}
