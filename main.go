package main

import (
	//	"encoding/binary"
	"fmt"
	"github.com/sstallion/go-hid"
	"io"
	"log"
	"math"

	//	"math"
	"strings"

	"github.com/fatih/color"
)

func main() {
	// show cursor
	defer func() {
		fmt.Print("\033[?25h")
	}()

	dev, err := hid.OpenFirst(0x10c4, 0xEA80)
	if err != nil {
		log.Fatalf("open device: %v", err)
	}

	_, err = dev.SendFeatureReport([]byte{0x41, 0x01})
	if err != nil {
		log.Fatalf("enable uart: %v", err)
	}

	// this refers to 9600, parity=NONE, 8 Bit, hardware Flow Control disabled, stop bits short
	_, err = dev.SendFeatureReport([]byte{0x50, 0x00, 0x00, 0x25, 0x80, 0x00, 0x00, 0x03, 0x00, 0x00})
	if err != nil {
		log.Fatalf("uart config: %v", err)
	}

	for {
		_, err = dev.Write([]byte{0x06, 0xab, 0xcd, 0x03, 0x5e, 0x01, 0xd9})
		if err != nil {
			log.Fatalf("start: %v", err)
		}

		message, err := readMessage(dev)
		if err != nil {
			log.Printf("readMessage: %v", err)
			continue
		}
		if message != nil {
			var hexMessage strings.Builder
			for _, val := range message {
				fmt.Fprintf(&hexMessage, "%02x ", val)
			}
			parseMessage(message)
			// log.Printf("message: %s ->  %s", hexMessage.String(), parseMessage(message))
		}
	}
}

func readMessage(dev io.Reader) ([]byte, error) {
	buf := make([]byte, 2)
	var message []byte
	for i := 0; ; i++ {
		_, err := dev.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("read: %v", err)
		}
		if buf[0] != 1 {
			return nil, fmt.Errorf("unexpected message:  %02x %02x", buf[0], buf[1])
		}
		val := buf[1]
		if (i == 0 && val != 0xab) || (i == 1 && val != 0xcd) {
			return nil, fmt.Errorf("invalid header, skipping message: i=%d, val=%02x", i, val)
		}
		if i == 2 {
			length := int(val)
			message = make([]byte, length)
		}
		if i > 2 && i < len(message)+2 {
			message[i-2] = val
		}
		if i == len(message)+2 {
			return message, nil
		}
	}
}

type Mode uint

const (
	Mode_VAC Mode = iota
	Mode_mVAC
	Mode_VDC
	Mode_mVDC
	Mode_Hz
	Mode_Percent
	Mode_Ohm
	Mode_Continuity
	Mode_Diode
	Mode_F
	Mode_0x0A
	Mode_0x0B
	Mode_uADC
	Mode_uAAC
	Mode_mADC
	Mode_mAAC
	Mode_ADC
	Mode_AAC
	Mode_hFE
	Mode_0x13
	Mode_NCV
	Mode_0x15
	Mode_0x16
	Mode_0x17
	Mode_LPF
	Mode_VAC_VDC
)

func (m Mode) String() string {
	str := []string{"V AC", "mV AC", "V DC", "mV DC", "Hz", "Percent", "Ohm", "Continuity", "Diode", "F", "0x0A", "0x0B", "uA DC", "uA AC", "mA DC", "mA AC", "A DC", "A AC", "hFE", "0x13", "NCV", "0x15", "0x16", "0x17", "LPF", "VAC_VDC"}
	if int(m) < len(str) {
		return str[m]
	} else {
		return fmt.Sprintf("0x%02x", uint(m))
	}
}

type Unit uint

const (
	Unit_None Unit = iota
	Unit_Volt
	Unit_Ampere
	Unit_Percent
	Unit_Ohm
	Unit_Hz
	Unit_Farad
)

func (u Unit) String() string {
	str := []string{"", "V", "A", "%", "Ω", "Hz", "F"}
	if int(u) < len(str) {
		return str[u]
	} else {
		return fmt.Sprintf("0x%02x", uint(u))
	}
}

type Message struct {
	Mode  Mode
	Unit  Unit
	Range byte
	Value float64
}

func (m Message) String() string {
	var str strings.Builder
	fmt.Fprintf(&str, "Mode:%s Range:%d", m.Mode, m.Range)
	return str.String()
}

func parseMessage(d []byte) (*Message, error) {
	var m Message
	m.Mode = Mode(d[1])
	m.Range = d[2] & 0x0F

	fmt.Print("\033[?25l")

	for i := range d {
		if i == 2 {
			color.New(color.FgGreen).Printf("%02x ", d[i])
		} else if i >= 4 && i <= 9 {
			color.New(color.FgYellow).Printf("%02x ", d[i])
		} else if i == 14 {
			color.New(color.FgBlue).Printf("%02x ", d[i])
		} else {
			fmt.Printf("%02x ", d[i])
		}
	}
	fmt.Print("\n")

	for i := 0; i < len(d); i++ {
		if i == 2 {
			color.New(color.FgGreen).Printf("%2d ", d[i]&0x0F)
		} else if i >= 4 && i <= 9 {
			if d[i] == 0x2e {
				color.New(color.FgYellow).Print(" , ")
			} else {
				color.New(color.FgYellow).Printf("%2d ", d[i]&0x0F)
			}
		} else if i == 14 {
			color.New(color.FgBlue).Printf("%2d ", d[i]&0x0F)
		} else {
			fmt.Printf("%2d ", d[i]&0x0F)
		}
	}

	fmt.Print("\n")

	var val int64
	comma := -1
	for i := 4; i < 10; i++ {
		if d[i] == 0x4f && d[i+1] == 0x4c {
			val = math.MaxInt64
			break
		} else if d[i] == 0x2e {
			comma = i
		} else {
			val *= 10
			val += int64(d[i] & 0x0F)
		}
	}

	var sign int64
	if d[14]&0x01 == 1 {
		sign = -1
	} else {
		sign = 1
	}

	if m.Mode == Mode_VAC_VDC {
		if d[14]&0x08 == 0x08 {
			m.Mode = Mode_VAC
		} else {
			m.Mode = Mode_VDC
		}
	}

	var factors []float64

	switch m.Mode {
	case Mode_VAC_VDC:
		m.Unit = Unit_Volt
		factors = []float64{1.0, 1.0, 1.0, 1.0}
	case Mode_VAC:
		m.Unit = Unit_Volt
		factors = []float64{1.0, 1.0, 1.0, 1.0}
	case Mode_mVAC:
		m.Unit = Unit_Volt
		factors = []float64{0.001, 0.001, 0.001, 0.001}
	case Mode_VDC:
		m.Unit = Unit_Volt
		factors = []float64{1.0, 1.0, 1.0, 1.0}
	case Mode_LPF:
		m.Unit = Unit_Volt
		factors = []float64{1.0, 1.0, 1.0, 1.0}
	case Mode_mVDC:
		factors = []float64{0.001, 0.001, 0.001, 0.001}
		m.Unit = Unit_Volt
	case Mode_Hz:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Hz
	case Mode_Percent:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Percent
	case Mode_Ohm:
		factors = []float64{1.0, 1000.0, 1000.0, 1000.0, 1000000.0, 1000000.0, 1000000.0}
		m.Unit = Unit_Ohm
	case Mode_Continuity:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Ohm
	case Mode_Diode:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Volt
	case Mode_F:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Farad
	case Mode_uADC:
		factors = []float64{1.0 / 1000000.0, 1.0 / 1000000.0, 1.0 / 1000000.0, 1.0 / 1000000.0}
		m.Unit = Unit_Ampere
	case Mode_uAAC:
		factors = []float64{1.0 / 1000000.0, 1.0 / 1000000.0, 1.0 / 1000000.0, 1.0 / 1000000.0}
		m.Unit = Unit_Ampere
	case Mode_mADC:
		factors = []float64{0.001, 0.001, 0.001, 0.001}
		m.Unit = Unit_Ampere
	case Mode_mAAC:
		factors = []float64{0.001, 0.001, 0.001, 0.001}
		m.Unit = Unit_Ampere
	case Mode_ADC:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Ampere
	case Mode_AAC:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_Ampere
	default:
		factors = []float64{1.0, 1.0, 1.0, 1.0}
		m.Unit = Unit_None
	}

	if int(m.Range) >= len(factors) {
		return nil, fmt.Errorf("invalid range (%d) for mode (%s)", m.Range, m.Mode.String())
	}
	factor := factors[m.Range]
	if val == math.MaxInt64 {
		m.Value = math.Inf(1)
	} else {
		m.Value = float64(sign*val) / math.Pow10(9-comma) * factor
	}

	fmt.Printf("                                                                      \r")
	fmt.Printf("range = %d, factor = %f, val = %.8f %s, mode = %s", m.Range, factor, m.Value, m.Unit.String(), m.Mode.String())
	fmt.Print("\r\033[2A")

	return &m, nil
}
