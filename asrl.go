// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

// Package asrl provides an Asynchronous Serial (ASRL) interface for
// controlling test equipment over serial ports using SCPI commands. It
// implements the VISA ASRL resource string format and serves as an instrument
// driver for the ivi and visa packages.
package asrl

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Device models a serial device and implements the ivi.Driver interface.
type Device struct {
	EndMark       byte
	HWHandshaking bool
	DelayTime     time.Duration
	port          serial.Port
}

// NewDevice opens a serial Device using the given VISA address resource string.
func NewDevice(address string) (*Device, error) {
	v, err := NewVisaResource(address)
	if err != nil {
		return nil, err
	}

	mode := &serial.Mode{
		BaudRate: v.baud,
		Parity:   v.parity,
		DataBits: v.dataBits,
		StopBits: v.stopBits,
	}
	port, err := serial.Open(v.address, mode)
	if err != nil {
		return nil, err
	}

	return &Device{
		port:          port,
		HWHandshaking: false,
		EndMark:       '\n',
		DelayTime:     70 * time.Millisecond,
	}, nil
}

// Write writes the given data to the network connection.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.port.Write(p)
}

// Read reads from the network connection into the given byte slice.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.port.Read(p)
}

// Close closes the underlying network connection.
func (d *Device) Close() error {
	time.Sleep(d.DelayTime)
	return d.port.Close()
}

// WriteString writes a string using the underlying network connection.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends the SCPI/ASCII command to the underlying network connection. A
// newline character is automatically added to the end of the string.
func (d *Device) Command(ctx context.Context, format string, a ...any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.HWHandshaking {
		if err := d.napIfDataSetNotReady(ctx); err != nil {
			return err
		}
	}
	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	cmd = strings.TrimSpace(cmd) + string(d.EndMark)
	if _, err := d.WriteString(cmd); err != nil {
		return err
	}
	time.Sleep(d.DelayTime)

	return nil
}

// Query writes the given string to the underlying network connection and
// returns a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(ctx context.Context, cmd string) (string, error) {
	if err := d.Command(ctx, cmd); err != nil {
		return "", err
	}
	return bufio.NewReader(d.port).ReadString('\n')
}

func isDSR(port serial.Port) (bool, error) {
	msb, err := port.GetModemStatusBits()
	if err != nil {
		return false, fmt.Errorf("getting modem status bits: %w", err)
	}
	return msb.DSR, nil
}

func (d *Device) napIfDataSetNotReady(ctx context.Context) error {
	// If I use 40 ms instead of 50 ms for the delay time, the Keysight E3631A DC
	// power supply will hang when sending commands/queries. Using 50 ms causes
	// the power supply to hang sometimes. I'm currently using 70 ms to be safe.
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ready, err := isDSR(d.port)
		if err != nil {
			return err
		}
		if ready {
			break
		}
		time.Sleep(d.DelayTime)
	}
	// Sleep a bit longer once the Data Set Ready is true. Without this, the
	// Keysight E3631A DC power supply will sometimes hang when sending
	// commands/queries.
	time.Sleep(d.DelayTime)
	return nil
}
