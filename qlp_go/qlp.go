package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

// QLP File Layout:
// [HEADER]
// 0x0004b U32 Magic Number 'IILQ'
// 0x0204b [ ... unknown purpose ... ]
//
// [WELL USE RECORDS x 96]
// 0x0004b U32 Unknown
// 0x0001b U8 Unknown
// 0x0001b U8 Unknown
// 0x0004b String ID
// 0x0100b String Name
// 0x0220b String Assay
// 0x0020b String Channel 1 Interpretation
// 0x0100b String Channel 1 Name
// 0x0020b String Channel 2 Interpretation
// 0x0100b String Channel 2 Name
//
// [WELL CONTENTS RECORDS x 96]
// 0x0004b U32 Unknown
// 0x0002b U16 Unknown
// 0x0002b U16 Unknown
// 0x0002b U16 Unknown
// 0x0004b String WellName
// 0x0084b [ ... unknown purpose ... ]
// VAR LEN String Catalog ID
// VAR LEN String Reagent Name
//
// [UNKNOWN PRE-MEASUREMENT BLOCK]
// ???
//
// [WELL MEASUREMENT RECORDS x 96]
// ???
// [MEASUREMENT DATA x Variable]
// 0x0004b U32 Measurement Point
// 0x0004b F32 Channel 1 Amplitude
// 0x0004b F32 Unknown (Possibly Width?)
// 0x0004b F32 Always 1.0f
// 0x0004b F32 Channel 2 Amplitude
// 0x0004b F32 Unknown (Possibly Width? Same value as above.)
// 0x0004b F32 Always 1.0f

type qlpHeader struct {
	Magic   [4]byte
	Unknown [0x204]byte
}

type Well struct {
	// A well is a set of:

	// Uses
	ID     string
	Assay  string
	CH1Use string
	CH2Use string

	// Contents
	CatalogID string
	Reagent   string

	// Measurements
	// (TODO)
}

func (w *Well) String() string {
	return fmt.Sprintf("Well %v [%s#%s] (%v); %v/%v", w.ID, w.CatalogID, w.Reagent, w.Assay, w.CH1Use, w.CH2Use)
}

type wellUseRecord struct {
	Unknown1 uint32
	Unknown2 byte
	Unknown3 byte
	ID       [4]byte
	Name     [0x100]byte
	Assay    [0x220]byte
	CH1Use   [0x20]byte
	CH1Name  [0x100]byte
	CH2Use   [0x20]byte
	CH2Name  [0x100]byte
}

type wellContentsRecord struct {
	Unknown1 uint32
	Unknown2 uint16
	Unknown3 uint16
	Unknown4 uint16
	ID       [4]byte
	Unknown5 [0x84]byte
}

func zts(b []byte) string {
	zb := bytes.IndexByte(b, 0)
	return string(b[0:zb])
}

type wellMeasurement struct {
	A                      uint32
	Ch1A, C, D, Ch2A, F, G float32
}

type QLPFile struct {
	header qlpHeader
	wells  [96]Well
	file   *os.File
}

func OpenQLPFile(path string) (*QLPFile, error) {
	failed := true
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(file)

	defer func() {
		// Only close the file if we're not returning a QLPFile instance.
		if failed {
			file.Close()
		}
	}()

	var hdr qlpHeader
	_, err = file.Seek(0x0, 0)
	br.Reset(file)
	if err != nil {
		return nil, err
	}

	err = binary.Read(br, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}

	var wells [96]Well

	var wellUseRecords [96]wellUseRecord
	err = binary.Read(br, binary.LittleEndian, &wellUseRecords)
	if err != nil {
		return nil, err
	}

	for i, wur := range wellUseRecords {
		wells[i] = Well{
			ID:     zts(wur.ID[:]),
			Assay:  zts(wur.Assay[:]),
			CH1Use: zts(wur.CH1Use[:]),
			CH2Use: zts(wur.CH2Use[:]),
		}
	}

	for i := 0; i < 96; i++ {
		var wcr wellContentsRecord
		err = binary.Read(br, binary.LittleEndian, &wcr)
		if err != nil {
			return nil, err
		}
		s, err := br.ReadString(0)
		if err != nil {
			return nil, err
		}
		wells[i].CatalogID = s[:len(s)-1]

		s, err = br.ReadString(0)
		if err != nil {
			return nil, err
		}
		wells[i].Reagent = s[:len(s)-1]
	}

	// TODO TODO TODO This seeks to A01's measurement block.
	_, err = file.Seek(0x247B7, 0)
	br.Reset(file)
	if err != nil {
		return nil, err
	}

	for _, well := range wells {
		logrus.Infof("%s", well.String())
	}

	failed = false
	return &QLPFile{
		header: hdr,
		wells:  wells,
		file:   file,
	}, nil

}

func (qf *QLPFile) ReadRecord() (wellMeasurement, error) {
	var rec wellMeasurement
	err := binary.Read(qf.file, binary.LittleEndian, &rec)
	return rec, err

}

func (qf *QLPFile) Close() error {
	qf.file.Close()
	return nil
}

func QLPtoCSV(qf *QLPFile, csvPath string) error {
	outf, _ := os.Create(csvPath)
	defer outf.Close()

	csvw := csv.NewWriter(outf)

	//var chs map[int]int
	i := 0
	csvw.Write([]string{"measurepoint", "ch1a", "c", "d", "ch2a", "f", "g"})
	for {
		rec, err := qf.ReadRecord()

		if rec.D != rec.G {
			logrus.Warning("Record D/G don't match (nominally 1.0)", rec)
		}

		if rec.C != rec.F {
			logrus.Warning("Record C/F don't match", rec)
		}

		if rec.G != 1.0 {
			break
		}

		if err != nil && err != io.EOF {
			return err
		}

		//fo, _ := qf.file.Seek(0, os.SEEK_CUR)
		//sFO := "0x" + strconv.FormatUint(uint64(fo), 16)
		sA := strconv.FormatUint(uint64(rec.A), 10)
		sB := strconv.FormatFloat(float64(rec.Ch1A), 'g', 8, 32)
		sC := strconv.FormatFloat(float64(rec.C), 'g', 8, 32)
		sD := strconv.FormatFloat(float64(rec.D), 'g', 8, 32)
		sE := strconv.FormatFloat(float64(rec.Ch2A), 'g', 8, 32)
		sF := strconv.FormatFloat(float64(rec.F), 'g', 8, 32)
		sG := strconv.FormatFloat(float64(rec.G), 'g', 8, 32)
		csvw.Write([]string{sA, sB, sC, sD, sE, sF, sG})
		i++

		if (i % 100000) == 0 {
			logrus.Infof("... %d ...", i)
		}

		if err == io.EOF {
			break
		}
	}
	csvw.Flush()
	logrus.Infof("--- %d measurements ---", i)
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

	inDir := args[1]
	err := filepath.Walk(args[1], func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			if path != inDir && !opts.Recurse {
				return filepath.SkipDir
			}
			return nil
		}

		logrus.Info(path)
		if filepath.Ext(path) != ".qlp" {
			return nil
		}

		qf, err := OpenQLPFile(path)
		if err != nil {
			return err
		}

		defer qf.Close()

		_, fileOnly := filepath.Split(path)
		csvPath := filepath.Join(opts.OutDir, fileOnly+".v16le.csv")

		err = QLPtoCSV(qf, csvPath)
		return err
	})
	if err != nil {
		logrus.Error(err)
	}
}
