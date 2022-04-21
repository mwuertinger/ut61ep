package ut61ep

import (
	"fmt"
	"github.com/sstallion/go-hid"
	"io"
	"math"
	"strings"
)

type Device interface {
	GetSerial() string
	ReadMessage() (*Message, error)
}

type Message struct {
	RawMessage []byte
	Mode       Mode
	Range      byte
	Unit       Unit
	Value      float64
}

func (m Message) String() string {
	var str strings.Builder
	fmt.Fprintf(&str, "Mode:%s Range:%d", m.Mode, m.Range)
	return str.String()
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

const (
	// CP2110 USB VID/PID
	usb_vid = 0x10C4
	usb_pid = 0xEA80
)

// Open the device with the specified serial number or the first device if serial number is empty.
func Open(serial string) (Device, error) {
	var dev *hid.Device
	var err error
	if serial == "" {
		dev, err = hid.OpenFirst(usb_vid, usb_pid)
	} else {
		dev, err = hid.Open(usb_vid, usb_pid, serial)
	}
	if err != nil {
		return nil, fmt.Errorf("open device: %v", err)
	}

	serial, err = dev.GetSerialNbr()
	if err != nil {
		return nil, fmt.Errorf("get serial: %v", err)
	}

	// See datasheet https://www.silabs.com/documents/public/application-notes/an434-cp2110-4-interface-specification.pdf
	// Enable UART (section 5.2 Get/Set UART Enable)
	_, err = dev.SendFeatureReport([]byte{0x41, 0x01})
	if err != nil {
		return nil, fmt.Errorf("enable uart: %v", err)
	}

	// UART config (section 6.3 Get/Set UART Config): 9600 baud, parity=NONE, 8 Bit, hardware Flow Control disabled, stop bits short
	_, err = dev.SendFeatureReport([]byte{0x50, 0x00, 0x00, 0x25, 0x80, 0x00, 0x00, 0x03, 0x00, 0x00})
	if err != nil {
		return nil, fmt.Errorf("uart config: %v", err)
	}

	return &device{hid: dev, serial: serial}, nil
}

type device struct {
	hid    io.ReadWriter
	serial string
}

var (
	requestData = []byte{0x06, 0xab, 0xcd, 0x03, 0x5e, 0x01, 0xd9}
)

func (d *device) GetSerial() string {
	return d.serial
}

func (d *device) ReadMessage() (*Message, error) {
	// request data
	_, err := d.hid.Write(requestData)
	if err != nil {
		return nil, fmt.Errorf("start: %v", err)
	}

	buf := make([]byte, 2)
	var message []byte
	for i := 0; len(message) == 0 || i <= len(message)+2; i++ {
		_, err := d.hid.Read(buf)
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
	}

	return parseMessage(message)
}

var factors = map[Mode][]float64{
	Mode_mVAC: {1.0 / 1000.0},
	Mode_mVDC: {1.0 / 1000.0},
	Mode_Ohm:  {1.0, 1000.0, 1000.0, 1000.0, 1000000.0},
	Mode_uADC: {1.0 / 1000000.0},
	Mode_uAAC: {1.0 / 1000000.0},
	Mode_mADC: {1.0 / 1000.0},
	Mode_mAAC: {1.0 / 1000.0},
}

var units = map[Mode]Unit{
	Mode_VAC_VDC:    Unit_Volt,
	Mode_VAC:        Unit_Volt,
	Mode_mVAC:       Unit_Volt,
	Mode_VDC:        Unit_Volt,
	Mode_LPF:        Unit_Volt,
	Mode_mVDC:       Unit_Volt,
	Mode_Hz:         Unit_Hz,
	Mode_Percent:    Unit_Percent,
	Mode_Ohm:        Unit_Ohm,
	Mode_Continuity: Unit_Ohm,
	Mode_Diode:      Unit_Volt,
	Mode_F:          Unit_Farad,
	Mode_uADC:       Unit_Ampere,
	Mode_uAAC:       Unit_Ampere,
	Mode_mADC:       Unit_Ampere,
	Mode_mAAC:       Unit_Ampere,
	Mode_ADC:        Unit_Ampere,
	Mode_AAC:        Unit_Ampere,
}

func parseMessage(d []byte) (*Message, error) {
	var m Message
	m.RawMessage = d
	m.Range = d[2] & 0x0F

	m.Mode = Mode(d[1])
	if m.Mode == Mode_VAC_VDC {
		if d[14]&0x08 == 0x08 {
			m.Mode = Mode_VAC
		} else {
			m.Mode = Mode_VDC
		}
	}

	if unit, ok := units[m.Mode]; ok {
		m.Unit = unit
	} else {
		m.Unit = Unit_None
	}

	var sign int64
	if d[14]&0x01 == 1 {
		sign = -1
	} else {
		sign = 1
	}

	factor := 1.0
	if modeFactors, ok := factors[m.Mode]; ok {
		i := int(m.Range)
		// pick the last factor in the list if range exceeds list length
		if i >= len(modeFactors) {
			i = len(modeFactors) - 1
		}
		factor = modeFactors[i]
	}

	const (
		valueStart = 4
		valueEnd   = 9
	)

	var val int64
	comma := valueEnd // set comma location to last value byte to divide by 1 below if comma is absent
	for i := valueStart; i <= valueEnd; i++ {
		// Over-Limit is indicated as 0x4F4C somewhere in the value
		if d[i] == 0x4F && d[i+1] == 0x4C {
			val = math.MaxInt64
			break
		} else if d[i] == 0x2E {
			// comma is marked with 0x2E
			comma = i
		} else {
			val *= 10
			val += int64(d[i] & 0x0F)
		}
	}

	if val == math.MaxInt64 {
		m.Value = math.Inf(1)
	} else {
		// shift value by comma location
		m.Value = float64(sign*val) / math.Pow10(valueEnd-comma) * factor
	}

	return &m, nil
}
