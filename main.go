package main

import (
	//	"encoding/binary"
	"fmt"
	"github.com/sstallion/go-hid"
	"log"
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

		message := readMessage(dev)
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

func readMessage(dev *hid.Device) []byte {
	buf := make([]byte, 2)
	var message []byte
	for i := 0; i < 100; i++ {
		_, err := dev.Read(buf)
		if err != nil {
			log.Printf("read: %v", err)
		}
		if buf[0] != 1 {
			log.Printf("unexpected message:  %02x %02x", buf[0], buf[1])
		}
		val := buf[1]
		if (i == 0 && val != 0xab) || (i == 1 && val != 0xcd) {
			log.Printf("invalid header, skipping message")
			return nil
		}
		if i == 2 {
			length := int(val)
			message = make([]byte, length)
		}
		if i > 2 && i < len(message)+2 {
			message[i-2] = val
		}
		if i == len(message)+2 {
			return message
		}
	}
	return nil
}

type Mode int

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
)

func (m Mode) String() string {
	str := []string{"V AC", "mV AC", "V DC", "mV DC", "Hz", "Percent", "Ohm", "Continuity", "Diode", "F", "0x0A", "0x0B", "uA DC", "uA AC", "mA DC", "mA AC", "A DC", "A AC", "hFE", "0x13", "NCV", "0x15", "0x16", "0x17", "LPF"}
	if int(m) < len(str) {
		return str[m]
	} else {
		return fmt.Sprintf("0x%02x", int(m))
	}
}

type Message struct {
	Mode  Mode
	Range byte
	Value float32
}

func (m Message) String() string {
	var str strings.Builder
	fmt.Fprintf(&str, "Mode:%s Range:%d", m.Mode, m.Range)
	return str.String()
}

func parseMessage(d []byte) *Message {
	var m Message
	m.Mode = Mode(d[1])
	m.Range = d[2]

	fmt.Print("\033[?25l")

	for i := range d {
		if i == 2 {
			color.New(color.FgGreen).Printf("%02x ", d[i])
		} else if i >= 4 && i <= 9 {
			color.New(color.FgYellow).Printf("%02x ", d[i])
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
		} else {
			fmt.Printf("%2d ", d[i]&0x0F)
		}
	}
	fmt.Print("\033[1A")
	fmt.Print("\r")

	//var digits []byte
	//for i, b := range d {
	//	digits = append()
	//}

	// fmt.Printf("%d%d%d%d\n", d[5]&0x0F, d[7]&0x0F, d[8]&0x0F, d[9]&0x0F)

	// fmt.Printf("length=%d\n", len(d))

	//	for i := 3; i < len(d)-3-4; i++ {
	//		bits := binary.BigEndian.Uint32(d[i : i+4])
	//		val := math.Float32frombits(bits)
	//		fmt.Printf("float[%d]=%f ", i, val)
	//		fmt.Printf("  int[%d]=%d ", i, bits)
	//	}
	//	fmt.Println()
	//for i := 3; i < len(d)-3-8; i++ {
	//	bits := binary.BigEndian.Uint64(d[i : i+8])
	//	val := math.Float64frombits(bits)
	//	fmt.Printf("val[%d]=%f ", i, val)
	//	// fmt.Printf("val[%d]=%d ", i, bits)
	//}
	//fmt.Println()

	return &m
}
