package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mwuertinger/ut61ep"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		fmt.Print("\033[?25h\n")
		os.Exit(0)
	}()

	dev, err := ut61ep.Open("")
	if err != nil {
		log.Fatalf("open: %v", err)
	}

	fmt.Printf("serial=%s\n", dev.GetSerial())

	for {
		message, err := dev.ReadMessage()
		if err != nil {
			log.Printf("readMessage: %v", err)
			continue
		}

		fmt.Print("\033[?25l") // hide cursor

		for i := range message.RawMessage {
			if i == 2 {
				color.New(color.FgGreen).Printf("%02x ", message.RawMessage[i])
			} else if i >= 4 && i <= 9 {
				color.New(color.FgYellow).Printf("%02x ", message.RawMessage[i])
			} else if i == 14 {
				color.New(color.FgBlue).Printf("%02x ", message.RawMessage[i])
			} else {
				fmt.Printf("%02x ", message.RawMessage[i])
			}
		}
		fmt.Print("\n")

		for i := 0; i < len(message.RawMessage); i++ {
			if i == 2 {
				color.New(color.FgGreen).Printf("%2d ", message.RawMessage[i]&0x0F)
			} else if i >= 4 && i <= 9 {
				if message.RawMessage[i] == 0x2e {
					color.New(color.FgYellow).Print(" , ")
				} else {
					color.New(color.FgYellow).Printf("%2d ", message.RawMessage[i]&0x0F)
				}
			} else if i == 14 {
				color.New(color.FgBlue).Printf("%2d ", message.RawMessage[i]&0x0F)
			} else {
				fmt.Printf("%2d ", message.RawMessage[i]&0x0F)
			}
		}

		fmt.Print("\n                                                                      \r")
		fmt.Printf("range = %d, val = %.8f %s, mode = %s", message.Range, message.Value, message.Unit.String(), message.Mode.String())
		fmt.Print("\r\033[2A") // move cursor to beginning of line and up 2 lines
	}
}
