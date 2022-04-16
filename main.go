package main

import (
	"encoding/hex"
	"github.com/sstallion/go-hid"
	"log"
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
			log.Printf("message: %s", hex.EncodeToString(message))
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
