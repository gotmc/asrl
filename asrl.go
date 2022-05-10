// Copyright (c) 2017-2021 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"bufio"
	"fmt"
	"strings"

	"go.bug.st/serial"
)

// Device models a serial device and implements the ivi.Driver interface.
type Device struct {
	port serial.Port
}

// NewDevice opens a serial Device using the given VISA address resource string.
func NewDevice(address string) (*Device, error) {
	var d Device
	v, err := NewVisaResource(address)
	if err != nil {
		return &d, err
	}

	mode := &serial.Mode{
		BaudRate: v.baud,
		Parity:   v.parity,
		DataBits: v.dataBits,
		StopBits: v.stopBits,
	}
	port, err := serial.Open(v.address, mode)
	if err != nil {
		return &d, err
	}

	d.port = port
	return &d, nil
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
	if err := d.port.ResetInputBuffer(); err != nil {
		return err
	}
	if err := d.port.ResetOutputBuffer(); err != nil {
		return err
	}
	return d.port.Close()
}

// WriteString writes a string using the underlying network connection.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends the SCPI/ASCII command to the underlying network connection. A
// newline character is automatically added to the end of the string.
func (d *Device) Command(format string, a ...interface{}) error {
	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	_, err := d.WriteString(strings.TrimSpace(cmd) + "\n")
	return err
}

// Query writes the given string to the underlying network connection and
// returns a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(cmd string) (string, error) {
	err := d.Command(cmd)
	if err != nil {
		return "", err
	}
	return bufio.NewReader(d.port).ReadString('\n')
}
