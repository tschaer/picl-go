/*
Driver file for N. Wirth's PICL compiler

Build with:
C:\> go install picl/piclc
Run:
C:\Data\Personal\go\bin> piclc file.asm
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"picl-go/PICL"
)

const Ver = "PICL compiler v1.0-alpha-2"

var (
   dump bool
)

func init() {
	flag.BoolVar(&dump, "d", false, "Dump listing output to console")
}

// Output an Intel HEX-record file
// format has 6 fields, all ASCII characters (2 chars per byte):
// : ll aaaa tt dd dd dd .... cc
// ll = 1 byte length, count only data bytes
// aaaa = 2 byte address
// tt = 1 byte type field (00 for normal, 01 for last record)
// dd = data bytes
func hexfile() {
   var byteH, byteL, checksum int
   var recs, lastrec int
   
   recs = PICL.Pc / 8
   lastrec = PICL.Pc % 8
   
   // Full records of 16 bytes (note each instruction is 2 bytes!)
   for i := 0; i < recs; i += 1 {
      fmt.Printf(":10 %.4X 00", i*16)
      checksum = 0
      for j := 0; j < 8; j += 1 {
         byteH = (PICL.Code[i*8+j] & 0xFF00) >> 8
         byteL = PICL.Code[i*8+j] & 0x00FF
         // Remember to byte swap!
         fmt.Printf(" %.2X %.2X", byteL, byteH)
         checksum = checksum + byteH + byteL
      }
      fmt.Printf(" %.2X\n", (^checksum+1) & 0x00FF)
   }
   
   // The last, partial record
   if lastrec > 0 {
      here := recs * 8
      fmt.Printf(":%.2X %.4X 00", lastrec*2, here*2)
      checksum = 0
      for i := here; i < PICL.Pc; i+= 1 {
         byteH = (PICL.Code[i] & 0xFF00) >> 8
         byteL = PICL.Code[i] & 0x00FF
         // Remember to byte swap!
         fmt.Printf(" %.2X %.2X", byteL, byteH)
         checksum = checksum + byteH + byteL
      }
      fmt.Printf(" %.2X\n", (^checksum+1) & 0x00FF)
   }
   
   // Terminating record
   fmt.Printf(":00 0000 01 FF\n")
}

func main() {

	fmt.Printf("%s\n\n", Ver)

	// Handle args
	flag.Parse()
	// Exit on error
	if !(len(flag.Args()) > 0) {
		fmt.Printf("Usage: piclc <flags> sourcefile.pcl\n")
		flag.PrintDefaults()
		return
	}

	// Compile source file. 
	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Compiling: %s\n", filename)
		PICL.Compile(bufio.NewReader(file))
	}
   
   // Handle options
   if dump {
		for addr := 0; addr < PICL.Pc; addr += 1 {
			fmt.Printf("%#.3x %#.4x\n", addr, PICL.Code[addr])
		}
	}
   
   // Output. Memory image is in PICL.Code
   hexfile()
   
}
