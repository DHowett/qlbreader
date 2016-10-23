package main

/*************************
 * research-quality code *
 *************************/

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Header struct {
	Magic   [4]byte
	Unknown [0xE2]byte
	NParam  uint16
}

type ParamRecord struct {
	Type   uint16
	Length uint16
}

type Record struct {
	CH1 uint16
	CH2 uint16
}

type CompensationRecord struct {
	Values [4]float32
}

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()

	var hdr Header
	f.Seek(0x0, 0)
	binary.Read(f, binary.LittleEndian, &hdr)
	fmt.Printf("Magic: %4s\nNumber of config records: %d\n", hdr.Magic, hdr.NParam)

	f.Seek(0x16A, 0)
	// read a zero-terminated string
	////// HACK -- assume it's Ch1,Ch2\0 (8 bytes)
	chS := make([]byte, 8)
	f.Read(chS)
	comps := make([]CompensationRecord, 4)
	binary.Read(f, binary.LittleEndian, &comps)
	for _, v := range comps {
		fmt.Printf("\n")
		fmt.Printf("    %9s | %9s\n", "Ch1", "Ch2")
		fmt.Printf("Ch1 %9f | %9f\n", v.Values[0], v.Values[1])
		fmt.Printf("Ch2 %9f | %9f\n\n", v.Values[2], v.Values[3])
	}

	f.Seek(0x1B2, 0)
	for i := uint16(0); i < hdr.NParam; i++ {
		var nameLen uint16
		binary.Read(f, binary.LittleEndian, &nameLen)

		bName := make([]byte, nameLen)
		f.Read(bName)
		name := string(bName[:nameLen-1])

		var prec ParamRecord
		binary.Read(f, binary.LittleEndian, &prec)
		switch prec.Type {
		case 0x0B:
			var val uint16
			binary.Read(f, binary.LittleEndian, &val)
			fmt.Printf("U16: %s = %d\n", name, val)
		case 0x02:
			val := make([]byte, prec.Length)
			f.Read(val)
			fmt.Printf("STR: %s = %s\n", name, string(val))
		default:
			fmt.Printf("UNK: %s (%x), bailing out\n", name, prec.Type)
			return
		}
	}

	outf, _ := os.Create(os.Args[1] + ".v16le.csv")
	defer outf.Close()
	f.Seek(0x47C+1, 0)
	//var chs map[int]int
	i := 0
	fmt.Fprintf(outf, "ch1,ch2\n")
	for {
		var rec Record
		err := binary.Read(f, binary.LittleEndian, &rec)
		fmt.Fprintf(outf, "%d,%d\n", rec.CH1, rec.CH2)
		i++
		if (i % 10000) == 0 {
			fmt.Fprintf(os.Stderr, "... %d ...\n", i)
		}

		if err == io.EOF {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "--- %d measurements ---", i)
}
