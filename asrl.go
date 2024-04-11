// Copyright (c) 2017-2024 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Device models a serial device and implements the ivi.Driver interface.
type Device struct {
	EndMark byte
	port    serial.Port
}

// NewDevice opens a serial Device using the given VISA address resource string.
func NewDevice(address string) (*Device, error) {
	v, err := NewVisaResource(address)
	if err != nil {
		return nil, err
	}
	log.Printf("Baud rate = %d", v.baud)
	log.Printf("Parity = %v", v.parity)
	log.Printf("Data bits = %d", v.dataBits)
	log.Printf("Stop bits = %v", v.stopBits)

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

	return &Device{port: port, EndMark: '\n'}, nil
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
	return d.port.Close()
}

// WriteString writes a string using the underlying network connection.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends the SCPI/ASCII command to the underlying network connection. A
// newline character is automatically added to the end of the string.
func (d *Device) Command(format string, a ...interface{}) error {
	// Debugging some timing issues that I believe are related to handshaking.
	showModemStatusBits(d.port)

	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	cmd = strings.TrimSpace(cmd) + string(d.EndMark)
	log.Printf("sending cmd: %s", cmd)
	_, err := d.WriteString(cmd)
	if err != nil {
		return err
	}

	// Debugging some timing issues that I believe are related to handshaking.
	showModemStatusBits(d.port)
	nap(200 * time.Millisecond)
	showModemStatusBits(d.port)
	return err
}

// Query writes the given string to the underlying network connection and
// returns a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(cmd string) (string, error) {
	msb, err := d.port.GetModemStatusBits()
	if err != nil {
		return "", err
	}
	log.Printf("%#v", msb)
	log.Printf("Getting ready to send cmd: %s", cmd)
	err = d.Command(cmd)
	if err != nil {
		return "", err
	}

	return bufio.NewReader(d.port).ReadString('\n')
}

func nap(duration time.Duration) {
	log.Printf("sleep for %s", duration)
	time.Sleep(duration)
}

func showModemStatusBits(port serial.Port) {
	msb, err := port.GetModemStatusBits()
	if err != nil {
		log.Printf("error getting modem status bits: %s", err)
	}
	log.Printf("DSR = %t", msb.DSR)
}

func getDSR(port serial.Port) (bool, error) {
	msb, err := port.GetModemStatusBits()
	if err != nil {
		return false, err
	}
	return msb.DSR, nil
}
