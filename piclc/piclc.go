/*
Driver file for N. Wirth's PICL compiler

Build with:
C:\> go install picl-go/piclc
Run:
C:\Data\Personal\go\bin> piclc file.pcl
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
   "io"
   "path/filepath"
	"os"
	"picl-go/PICL"
)

const Ver = "PICL compiler v1.0-beta-1"

var (
   dump bool
   list bool
)

func init() {
	flag.BoolVar(&dump, "d", false, "Dump program memory image to console")
   flag.BoolVar(&list, "l", false, "Generate listing file")
}

// Output an Intel HEX-record file
// format has 6 fields, all ASCII characters (2 chars per byte):
// : ll aaaa tt dd dd dd .... cc
// ll = 1 byte length, count only data bytes
// aaaa = 2 byte address
// tt = 1 byte type field (00 for normal, 01 for last record)
// dd = data bytes
// cc = checksum
func hexfile(f io.Writer) {
   var byteH, byteL, checksum int
   var recs, lastrec, addr int
   
   recs = PICL.Pc / 8
   lastrec = PICL.Pc % 8
   
   // Full records of 16 bytes (note each instruction is 2 bytes!)
   for i := 0; i < recs; i += 1 {
      addr = i*16
      fmt.Fprintf(f, ":10%.4X00", addr)
      checksum = 0x10 + ((addr & 0xFF00) >> 8) + (addr & 0x00FF)
      for j := 0; j < 8; j += 1 {
         byteH = (PICL.Code[i*8+j] & 0xFF00) >> 8
         byteL = PICL.Code[i*8+j] & 0x00FF
         // Remember to byte swap!
         fmt.Fprintf(f, "%.2X%.2X", byteL, byteH)
         checksum = checksum + byteH + byteL
      }
      fmt.Fprintf(f, "%.2X\n", (^(checksum & 0x00FF) + 1) & 0x00FF)
   }
   
   // The last, partial record
   if lastrec > 0 {
      here := recs * 8
      addr = here * 2
      fmt.Fprintf(f, ":%.2X%.4X00", lastrec*2, addr)
      checksum = lastrec*2 + ((addr & 0xFF00) >> 8) + (addr & 0x00FF)
      for i := here; i < PICL.Pc; i+= 1 {
         byteH = (PICL.Code[i] & 0xFF00) >> 8
         byteL = PICL.Code[i] & 0x00FF
         // Remember to byte swap!
         fmt.Fprintf(f, "%.2X%.2X", byteL, byteH)
         checksum = checksum + byteH + byteL
      }
      fmt.Fprintf(f, "%.2X\n", (^(checksum & 0x00FF) + 1) & 0x00FF)
   }
   
   // Terminating record
   fmt.Fprintf(f, ":00000001FF\n")
   
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

	// Open source file
	filename := flag.Arg(0)
   if filepath.Ext(filename) != ".pcl" {
      fmt.Printf("Source file must end in .pcl\n")
      return
   }
	file, err := os.Open(filename)
	
   // Compile
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
   
   // Output on successful compile
   if !PICL.Err {
      fname := filepath.Base(filename)
      froot := fname[:len(fname)-4]
      h, _ := os.Create(froot+".hex")
      hexfile(h)
      h.Close()
      if list {
         l, _ := os.Create(froot+".lst")
         PICL.Decode(l)
         l.Close()
      }
   }
   
}
