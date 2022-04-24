package main

import (
	"log"

	"github.com/mwuertinger/ut61ep"
)

func main() {
	dev, err := ut61ep.Open("")
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	message, err := dev.ReadMessage()
	if err != nil {
		log.Fatalf("readMessage: %v", err)
	}
	log.Printf("%f %s", message.Value, message.Unit.String())
}
