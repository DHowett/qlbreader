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

type paramRecord struct {
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

type QLBFile struct {
	header Header
	comps  []CompensationRecord
	config map[string]interface{}

	file *os.File
}

func OpenQLBFile(path string) (*QLBFile, error) {
	failed := true
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Only close the file if we're not returning a QLBFile instance.
		if failed {
			file.Close()
		}
	}()

	var hdr Header
	_, err = file.Seek(0x0, 0)
	if err != nil {
		return nil, err
	}

	err = binary.Read(file, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(0x16A, 0)
	if err != nil {
		return nil, err
	}

	// read a zero-terminated string
	////// HACK -- assume it's Ch1,Ch2\0 (8 bytes)
	chS := make([]byte, 8)
	_, err = file.Read(chS)
	if err != nil {
		return nil, err
	}

	comps := make([]CompensationRecord, 4)
	err = binary.Read(file, binary.LittleEndian, &comps)
	if err != nil {
		return nil, err
	}

	/*
		for _, v := range comps {
			fmt.Printf("\n")
			fmt.Printf("    %9s | %9s\n", "Ch1", "Ch2")
			fmt.Printf("Ch1 %9f | %9f\n", v.Values[0], v.Values[1])
			fmt.Printf("Ch2 %9f | %9f\n\n", v.Values[2], v.Values[3])
		}
	*/

	_, err = file.Seek(0x1B2, 0)
	if err != nil {
		return nil, err
	}

	configRecords := make(map[string]interface{}, hdr.NParam)
	for i := uint16(0); i < hdr.NParam; i++ {
		var nameLen uint16
		err = binary.Read(file, binary.LittleEndian, &nameLen)
		if err != nil {
			return nil, err
		}

		bName := make([]byte, nameLen)
		_, err = file.Read(bName)
		if err != nil {
			return nil, err
		}

		name := string(bName[:nameLen-1])

		var prec paramRecord
		err = binary.Read(file, binary.LittleEndian, &prec)
		if err != nil {
			return nil, err
		}

		switch prec.Type {
		case 0x0B:
			var val uint16
			err = binary.Read(file, binary.LittleEndian, &val)
			if err != nil {
				return nil, err
			}

			configRecords[name] = val
		case 0x02:
			val := make([]byte, prec.Length)
			_, err = file.Read(val)
			if err != nil {
				return nil, err
			}

			configRecords[name] = string(val)
		default:
			return nil, fmt.Errorf("unknown config record (name=%s, type=%x)", name, prec.Type)
		}
	}

	/// 0x47D: start of records
	/// --- might be contingent upon header size (!) ---
	_, err = file.Seek(0x47D, 0)
	if err != nil {
		return nil, err
	}

	failed = false
	return &QLBFile{
		header: hdr,
		comps:  comps,
		config: configRecords,
		file:   file,
	}, nil

}

func (qf *QLBFile) ReadRecord() (Record, error) {
	var rec Record
	err := binary.Read(qf.file, binary.LittleEndian, &rec)
	return rec, err

}

func (qf *QLBFile) Close() error {
	qf.file.Close()
	return nil
}

func QLBtoCSV(qf *QLBFile, csvPath string) error {
	outf, _ := os.Create(csvPath)
	defer outf.Close()

	//var chs map[int]int
	i := 0
	fmt.Fprintf(outf, "ch1,ch2\n")
	for {
		rec, err := qf.ReadRecord()
		if err != nil && err != io.EOF {
			return err
		}

		fmt.Fprintf(outf, "%d,%d\n", rec.CH1, rec.CH2)
		i++

		if (i % 10000) == 0 {
			fmt.Fprintf(os.Stderr, "... %d ...\n", i)
		}

		if err == io.EOF {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "--- %d measurements ---\n", i)
	return nil
}

func main() {
	qf, err := OpenQLBFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	defer qf.Close()

	csvPath := os.Args[1] + ".v16le.csv"
	err = QLBtoCSV(qf, csvPath)
	if err != nil {
		panic(err)
	}
}
