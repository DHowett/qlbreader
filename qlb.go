package main

/*************************
 * research-quality code *
 *************************/

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jessevdk/go-flags"
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

func (qf *QLBFile) GetCompensationMatrix(idx int) (*CompensationRecord, error) {
	if idx < 0 || idx >= len(qf.comps) {
		return nil, errors.New("out of range")
	}
	return &qf.comps[idx], nil
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

	csvw := csv.NewWriter(outf)

	//var chs map[int]int
	i := 0
	csvw.Write([]string{"ch1", "ch2"})
	for {
		rec, err := qf.ReadRecord()
		if err != nil && err != io.EOF {
			return err
		}

		sCH1 := strconv.FormatUint(uint64(rec.CH1), 10)
		sCH2 := strconv.FormatUint(uint64(rec.CH2), 10)
		csvw.Write([]string{sCH1, sCH2})
		i++

		if (i % 100000) == 0 {
			fmt.Fprintf(os.Stderr, "... %d ...\n", i)
		}

		if err == io.EOF {
			break
		}
	}
	csvw.Flush()
	fmt.Fprintf(os.Stderr, "--- %d measurements ---\n", i)
	return nil
}

type Options struct {
	OutDir  string `short:"o" long:"out" description:"output directory" default:"out"`
	Recurse bool   `short:"r" description:"recurse into subdirectories" default:"false"`
}

func main() {
	var opts Options
	args, _ := flags.ParseArgs(&opts, os.Args)

	os.MkdirAll(opts.OutDir, os.FileMode(0755))

	compFile, err := os.Create(filepath.Join(opts.OutDir, "comps.csv"))
	if err != nil {
		panic(err)
	}
	defer compFile.Close()
	compCSVWriter := csv.NewWriter(compFile)
	compCSVWriter.Write([]string{"ch1", "ch2"})

	inDir := args[1]
	filepath.Walk(args[1], func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			if path != inDir && !opts.Recurse {
				return filepath.SkipDir
			}
			return nil
		}

		fmt.Println(path)
		if filepath.Ext(path) != ".qlb" {
			return nil
		}

		qf, err := OpenQLBFile(path)
		if err != nil {
			return err
		}

		defer qf.Close()

		if compCSVWriter != nil {
			comps, _ := qf.GetCompensationMatrix(0)
			s := make([]string, len(comps.Values))
			for i, v := range comps.Values {
				s[i] = strconv.FormatFloat(float64(v), 'g', 9, 32)
			}
			compCSVWriter.Write(s[0:2])
			compCSVWriter.Write(s[2:4])
			compCSVWriter.Flush()
			compCSVWriter = nil
		}

		_, fileOnly := filepath.Split(path)
		csvPath := filepath.Join(opts.OutDir, fileOnly+".v16le.csv")
		err = QLBtoCSV(qf, csvPath)
		return err
	})
}
