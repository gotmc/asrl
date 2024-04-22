// Copyright (c) 2020–2024 The prologix developers. All rights reserved.
// Project site: https://github.com/gotmc/prologix
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gotmc/asrl"
)

var (
	serialPort string
	baudRate   int
)

func init() {

	// Get serial port for DS345.
	flag.StringVar(
		&serialPort,
		"port",
		"/dev/tty.usbserial-PX8X3YR6",
		"Serial port for DS345",
	)

	flag.IntVar(&baudRate, "baud", 9600, "Baud rate")
}

func main() {
	// Parse the flags
	flag.Parse()

	// Open virtual comm port.
	address := fmt.Sprintf("ASRL::%s::%d::8N2::INSTR", serialPort, baudRate)
	log.Printf("VISA Address = %s", address)
	dev, err := asrl.NewDevice(address)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// Query the identification of the function generator.
	idn, err := dev.Query("*idn?")
	if err != nil && err != io.EOF {
		log.Fatalf("error querying serial port: %s", err)
	}
	log.Printf("query idn = %s", idn)

	// Send commands to the function generator required to create a coded carrier
	// operating at 100 Hz with 400 ms on time and 200 ms off time.
	cmds := []string{
		"MENA0",      // Disable modulation
		"FUNC 0",     // Set output function to sine wave
		"FREQ 100.0", // Set frequency to 100 Hz
		"AMPL 0.5VP", // Set the amplitude to 0.5 Vpp
		"OFFS 0.0",   // Set the offset to 0 Vdc
		"PHSE 0.0",   // Set the phase to 0 degrees
		"BCNT 40",    // Set burst count to 40 (rounded to nearest integer)
		"TSRC1",      // Use internal trigger
		"TRAT 1.667", // Set the trigger rate to 1.667 Hz
		"MTYP5",      // Set the modulation to burst
		"MENA1",      // Enable modulation
	}
	for _, cmd := range cmds {
		log.Printf("Sending command: %s", cmd)
		err = dev.Command(cmd)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
