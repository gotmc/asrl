// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl_test

import (
	"context"
	"fmt"
	"log"

	"github.com/gotmc/asrl"
)

func Example() {
	// Open a serial device using a VISA resource string.
	dev, err := asrl.NewDevice("ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = dev.Close() }()

	ctx := context.Background()

	// Query the instrument identification.
	idn, err := dev.Query(ctx, "*IDN?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(idn)

	// Send a SCPI command.
	if err := dev.Command(ctx, "OUTP ON"); err != nil {
		log.Fatal(err)
	}
}

func Example_withOptions() {
	// Open a device with functional options.
	dev, err := asrl.NewDevice(
		"ASRL::/dev/tty.usbserial-PX8X3YR6::9600::8N2::INSTR",
		asrl.WithHWHandshaking(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = dev.Close() }()

	ctx := context.Background()

	// With hardware handshaking enabled, Command polls DSR before writing.
	if err := dev.Command(ctx, "SYST:REM"); err != nil {
		log.Fatal(err)
	}
}
