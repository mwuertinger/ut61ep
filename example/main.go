// Copyright (C) 2022 Michael Würtinger
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

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
