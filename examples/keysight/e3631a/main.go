// Copyright (c) 2024 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/gotmc/asrl"
)

var (
	debugLevel uint
	serialPort string
)

func init() {
	// Get the debug level from CLI flag.
	const (
		defaultLevel = 1
		debugUsage   = "USB debug level"
	)
	flag.UintVar(&debugLevel, "debug", defaultLevel, debugUsage)
	flag.UintVar(&debugLevel, "d", defaultLevel, debugUsage+" (shorthand)")

	// Get Virtual COM Port (VCP) serial port for Prologix.
	flag.StringVar(
		&serialPort,
		"port",
		"/dev/tty.usbserial-PX8X3YR6",
		"Serial port for Keysight E3631A",
	)
}

func main() {
	// Parse the flags
	flag.Parse()

	// Open virtual comm port.
	address := fmt.Sprintf("ASRL::%s::9600::8N2::INSTR", serialPort)
	log.Printf("VISA Address = %s", address)
	dev, err := asrl.NewDevice(address)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// Query the identification of the function generator.
	log.Println("Query the identification.")
	idn, err := dev.Query("*IDN?\r\n")
	if err != nil && err != io.EOF {
		log.Fatalf("error querying serial port: %s", err)
	}
	log.Printf("query idn = %s", idn)

	cmds := []string{
		"SYST:REM",
		"*RST",
		"*CLS",
		"appl p6v,1.7,1.3",
		"outp on",
	}

	for _, cmd := range cmds {
		log.Printf("ex app sending cmd = %s", cmd)
		if err = dev.Command(cmd); err != nil {
			log.Printf("Received error: %s", err)
			log.Fatal(err)
		}
		log.Printf("ex app sent cmd = %s", cmd)
	}

	// Query the voltage output
	log.Println("Query the voltage output.")
	vc, err := dev.Query("appl? p6v")
	if err != nil {
		log.Fatalf("error querying serial port: %s", err)
	}
	log.Printf("voltage, current = %s", vc)

	// Query the output state
	state, err := dev.Query("OUTP:STAT?")
	if err != nil && err != io.EOF {
		log.Fatalf("error querying serial port: %s", err)
	}
	log.Printf("output state = %s", state)
}