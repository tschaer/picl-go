/*
PICS.go: The PICL Scanner
Texts.Read(R, ch) -> ch, err = r.ReadByte()
*/

package PICS

import (
	"bufio"
)

var (
   ch byte
   r *bufio.Reader
)

/* Scanner init */
func Init(reader *bufio.Reader) {
	r = reader
	ch, _ = r.ReadByte()
}

/* Return next symbol in input */
func Get(sym *int) {
}
