package wd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type WriteDebugger interface {
	io.WriteSeeker
	Printf(format string, a ...interface{}) (int, error)
	SetComparison(r io.ReadSeeker)
	CompareRead() (int64, byte, error)
}

type WriteDebug struct {
	index int64
	Input io.ReadSeeker
}

func (wd *WriteDebug) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}

func (wd *WriteDebug) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (wd *WriteDebug) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (wd *WriteDebug) SetComparison(r io.ReadSeeker) {
	wd.Input = r
}

func (wd *WriteDebug) CompareRead() (int64, byte, error) {
	wd.index++
	if wd.Input == nil {
		return wd.index, 0, fmt.Errorf("no input set to compare")
	}
	data := make([]byte, 1)
	_, err := wd.Input.Read(data)
	if err != nil {
		return wd.index, 0, err
	}
	return wd.index, data[0], nil
}

func PrintWrite(w interface{}, order binary.ByteOrder, data interface{}, format string, a ...interface{}) error {
	switch out := w.(type) {
	case WriteDebugger:
		out.Printf(format, a...)
		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.LittleEndian, data)
		if err != nil {
			return err
		}

		for _, b := range buf.Bytes() {
			index, ab, err := out.CompareRead()
			if err != nil {
				return fmt.Errorf("compareread: %w", err)
			}
			if ab != b {
				return fmt.Errorf("compare at %d expected 0x%x, got 0x%x", index, ab, b)
			}
		}

		/*in := hex.Dump(buf.Bytes())
		outDump := ""
		matches := pat.FindAllStringSubmatch(in, -1)
		for _, subs := range matches {
			for i, sub := range subs {
				if i == 0 {
					continue
				}

				outDump += fmt.Sprintf("0x%s,\n", strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(sub, "  ", " ")), " ", ", 0x"))
			}
		}

		out.Printf("\n%s", outDump)
		*/
		return binary.Write(out, binary.LittleEndian, data)
	case io.WriteSeeker:
		return binary.Write(out, binary.LittleEndian, data)
	default:
		return fmt.Errorf("invalid printWrite destination type")
	}
}
