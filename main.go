package main

import (
	"encoding/hex"
	"fmt"
	"github.com/sstallion/go-hid"
	"log"
	"strings"
)

func main() {
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
			log.Printf("message: %s -> %s", hex.EncodeToString(message), parseMessage(message))
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
)

func (m Mode) String() string {
	str := []string{"V AC", "mV AC", "V DC", "mV DC", "Hz", "Percent", "Ohm", "Continuity", "Diode", "F", "0x0A", "0x0B", "uA DC", "uA AC", "mA DC", "mA AC", "A DC", "A AC", "hFE", "0x13", "NCV"}
	if int(m) < len(str) {
		return str[m]
	} else {
		return "-"
	}
}

type Message struct {
	Mode Mode
}

func (m Message) String() string {
	var str strings.Builder
	fmt.Fprintf(&str, "Mode:%s", m.Mode)
	return str.String()
}

func parseMessage(d []byte) *Message {
	var m Message
	m.Mode = Mode(d[1])

	return &m
}
