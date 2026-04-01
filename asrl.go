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
	ReadTimeout   time.Duration
	port          serial.Port
	reader        *bufio.Reader
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

	d := &Device{
		port:          port,
		reader:        bufio.NewReader(port),
		HWHandshaking: false,
		EndMark:       '\n',
		DelayTime:     70 * time.Millisecond,
		ReadTimeout:   5 * time.Second,
	}
	if err := port.SetReadTimeout(d.ReadTimeout); err != nil {
		port.Close()
		return nil, fmt.Errorf("setting read timeout: %w", err)
	}
	return d, nil
}

// Read reads from the serial port into the given byte slice.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.port.Read(p)
}

// Write writes the given data to the serial port.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.port.Write(p)
}

// ReadContext reads from the serial port into the given byte slice with context
// support. If the context is cancelled before the read completes, ReadContext
// sets a short timeout to unblock the read, waits for the goroutine to finish,
// resets the reader, and returns the context error.
func (d *Device) ReadContext(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	type result struct {
		n   int
		err error
	}
	ch := make(chan result, 1)
	go func() {
		n, err := d.port.Read(p)
		ch <- result{n, err}
	}()

	select {
	case <-ctx.Done():
		// Set a short read timeout to unblock the goroutine stuck on Read,
		// then wait for it to finish so we don't leak it.
		_ = d.port.SetReadTimeout(1 * time.Millisecond)
		<-ch
		_ = d.port.SetReadTimeout(d.ReadTimeout)
		return 0, ctx.Err()
	case r := <-ch:
		return r.n, r.err
	}
}

// WriteContext writes the given data to the serial port with context support.
// If the context is cancelled before the write completes, WriteContext returns
// the context error.
func (d *Device) WriteContext(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	type result struct {
		n   int
		err error
	}
	ch := make(chan result, 1)
	go func() {
		n, err := d.port.Write(p)
		ch <- result{n, err}
	}()

	select {
	case <-ctx.Done():
		// Set a short read timeout to unblock the goroutine, then wait for
		// it to finish so we don't leak it.
		_ = d.port.SetReadTimeout(1 * time.Millisecond)
		<-ch
		_ = d.port.SetReadTimeout(d.ReadTimeout)
		return 0, ctx.Err()
	case r := <-ch:
		return r.n, r.err
	}
}

// Close closes the underlying serial port.
func (d *Device) Close() error {
	time.Sleep(d.DelayTime)
	return d.port.Close()
}

// WriteString writes a string to the serial port. An endmark character, such
// as a newline, is not automatically added to the end of the string.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends a SCPI/ASCII command to the serial port. An endmark
// charachter, such as newline, is automatically added to the end of the
// string.
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
	if len(a) > 0 {
		cmd = fmt.Sprintf(format, a...)
	}
	cmd = strings.TrimSpace(cmd) + string(d.EndMark)
	if _, err := d.WriteString(cmd); err != nil {
		return err
	}
	time.Sleep(d.DelayTime)

	return nil
}

// Query writes the given SCPI/ASCII command to the serial port and returns the
// response string. The device's endmark character (newline by default) is
// automatically added to the query command. The string returned is not
// stripped of any whitespace. The context is used for cancellation; if the
// context is cancelled while waiting for a response, Query returns the context
// error.
func (d *Device) Query(ctx context.Context, cmd string) (string, error) {
	if err := d.Command(ctx, "%s", cmd); err != nil {
		return "", err
	}

	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := d.reader.ReadString(d.EndMark)
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		// Set a short read timeout to unblock the goroutine stuck on
		// ReadString, then wait for it to finish so we don't leak it or
		// race on the bufio.Reader.
		_ = d.port.SetReadTimeout(1 * time.Millisecond)
		<-ch
		_ = d.port.SetReadTimeout(d.ReadTimeout)
		d.reader.Reset(d.port)
		return "", ctx.Err()
	case r := <-ch:
		return r.s, r.err
	}
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
	deadline := time.Now().Add(d.ReadTimeout)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("asrl: DSR not ready after %s", d.ReadTimeout)
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
