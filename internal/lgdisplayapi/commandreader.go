package lgdisplayapi

import (
	"bufio"
	"fmt"
	"io"
)

const (
	cr = '\x0d'
)

type CommandReader struct {
	r *bufio.Reader
}

func newCommandReader(r io.Reader) *CommandReader {
	return &CommandReader{r: bufio.NewReader(r)}
}

func (cmdr *CommandReader) Next() (*command, error) {
	data, err := cmdr.r.ReadBytes(cr)
	if err != nil {
		return nil, err
	}
	return parseCommand(data)
}

type LGCommand string

const (
	LGPower       LGCommand = "ka"
	LGInput       LGCommand = "xb"
	LGSerial      LGCommand = "fy"
	LGVersion     LGCommand = "fz"
	LGTemperature LGCommand = "dn"
)

type LGValue string

const (
	// PowerValues
	PowerOff     LGValue = "00"
	PowerOn      LGValue = "01"
	PowerRestart LGValue = "02"
	// InputValues
	InputHDMI1 LGValue = "A0"
	InputHDMI2 LGValue = "A1"
	InputHDMI3 LGValue = "A7"
	InputDP1   LGValue = "D0"
)

type command struct {
	Cmd   LGCommand
	Id    int
	Value LGValue
}

func parseCommand(data []byte) (*command, error) {
	var s string
	for _, c := range data {
		s += string(rune(c))
	}
	var cmd command
	switch x := s[:1]; x {
	case string(LGPower), string(LGInput):
		cmd.Cmd = LGCommand(x)
	default:
		return nil, fmt.Errorf("unKnown command")
	}

	return &cmd, nil
}
