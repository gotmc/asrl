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
	ctx := context.Background()

	// Open a serial device using a VISA resource string.
	dev, err := asrl.NewDevice(ctx, "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

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

func ExampleNewDevice() {
	ctx := context.Background()
	dev, err := asrl.NewDevice(ctx, "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	fmt.Println("opened serial device")
}

func ExampleNewDevice_withOptions() {
	ctx := context.Background()
	dev, err := asrl.NewDevice(ctx,
		"ASRL::/dev/tty.usbserial-PX8X3YR6::9600::8N2::INSTR",
		asrl.WithHWHandshaking(true),
		asrl.WithDelayTime(100),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	fmt.Println("opened serial device with hardware handshaking")
}

func ExampleNewVisaResource() {
	v, err := asrl.NewVisaResource("ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v.InterfaceType())
	fmt.Println(v.Address())
	fmt.Println(v.Baud())

	// Output:
	// ASRL
	// /dev/tty.usbserial-PX484GRU
	// 9600
}

func ExampleDevice_Command() {
	ctx := context.Background()
	dev, err := asrl.NewDevice(ctx, "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// Command sends a SCPI command with an auto-appended endmark character.
	if err := dev.Command(ctx, "*RST"); err != nil {
		log.Fatal(err)
	}

	// Command supports fmt.Sprintf-style formatting.
	if err := dev.Command(ctx, "VOLT %f", 5.0); err != nil {
		log.Fatal(err)
	}
}

func ExampleDevice_Query() {
	ctx := context.Background()
	dev, err := asrl.NewDevice(ctx, "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR")
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// Query sends a command and reads the response.
	idn, err := dev.Query(ctx, "*IDN?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(idn)
}
