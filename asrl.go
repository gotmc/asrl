// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

// ErrDSRNotReady is returned when the Data Set Ready signal is not asserted
// within the ReadTimeout period.
var ErrDSRNotReady = errors.New("asrl: DSR not ready")

// Device models a serial device and implements the ivi.Transport interface.
type Device struct {
	endMark       byte
	hwHandshaking bool
	delayTime     time.Duration
	readTimeout   time.Duration
	port          serial.Port
	reader        *bufio.Reader
}

// EndMark returns the end-of-message byte used by Command and Query.
func (d *Device) EndMark() byte { return d.endMark }

// SetEndMark sets the end-of-message byte used by Command and Query.
func (d *Device) SetEndMark(b byte) { d.endMark = b }

// HWHandshaking returns whether hardware handshaking (DSR polling) is enabled.
func (d *Device) HWHandshaking() bool { return d.hwHandshaking }

// SetHWHandshaking enables or disables hardware handshaking (DSR polling).
func (d *Device) SetHWHandshaking(enabled bool) { d.hwHandshaking = enabled }

// DelayTime returns the delay between serial operations.
func (d *Device) DelayTime() time.Duration { return d.delayTime }

// SetDelayTime sets the delay between serial operations.
func (d *Device) SetDelayTime(t time.Duration) { d.delayTime = t }

// ReadTimeout returns the read timeout on the serial port.
func (d *Device) ReadTimeout() time.Duration { return d.readTimeout }

// SetReadTimeout sets the read timeout on the serial port.
func (d *Device) SetReadTimeout(t time.Duration) { d.readTimeout = t }

// DeviceOption is a functional option for configuring a Device.
type DeviceOption func(*Device)

// WithEndMark sets the end-of-message byte used by Command and Query.
func WithEndMark(b byte) DeviceOption {
	return func(d *Device) {
		d.endMark = b
	}
}

// WithHWHandshaking enables or disables hardware handshaking (DSR polling).
func WithHWHandshaking(enabled bool) DeviceOption {
	return func(d *Device) {
		d.hwHandshaking = enabled
	}
}

// WithDelayTime sets the delay between serial operations.
func WithDelayTime(t time.Duration) DeviceOption {
	return func(d *Device) {
		d.delayTime = t
	}
}

// WithReadTimeout sets the read timeout on the serial port.
func WithReadTimeout(t time.Duration) DeviceOption {
	return func(d *Device) {
		d.readTimeout = t
	}
}

// NewDevice opens a serial Device using the given VISA address resource string.
// The context is checked before opening the serial port. Optional DeviceOption
// values can be provided to override the default settings for EndMark,
// HWHandshaking, DelayTime, and ReadTimeout.
func NewDevice(ctx context.Context, address string, opts ...DeviceOption) (*Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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
		hwHandshaking: false,
		endMark:       '\n',
		delayTime:     70 * time.Millisecond,
		readTimeout:   5 * time.Second,
	}
	for _, opt := range opts {
		opt(d)
	}
	if err := port.SetReadTimeout(d.readTimeout); err != nil {
		_ = port.Close()
		return nil, fmt.Errorf("setting read timeout: %w", err)
	}
	return d, nil
}

// Close closes the underlying serial port.
func (d *Device) Close() error {
	time.Sleep(d.delayTime)
	return d.port.Close()
}

// Read reads from the serial port into the given byte slice.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.port.Read(p)
}

// Write writes the given data to the serial port.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.port.Write(p)
}

// WriteString writes a string to the serial port. An endmark character, such
// as a newline, is not automatically added to the end of the string.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// ReadBinary reads binary data from the serial port without terminator
// interpretation. If the context is canceled before the read completes,
// ReadBinary sets a short timeout to unblock the read, waits for the goroutine
// to finish, and returns the context error.
func (d *Device) ReadBinary(ctx context.Context, p []byte) (int, error) {
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
		_ = d.port.SetReadTimeout(d.readTimeout)
		return 0, ctx.Err()
	case r := <-ch:
		return r.n, r.err
	}
}

// WriteBinary writes binary data to the serial port without adding a
// terminator. If the context is already canceled before the write begins,
// WriteBinary returns the context error. Serial writes are typically
// non-blocking, so no goroutine-based cancellation is needed.
func (d *Device) WriteBinary(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	return d.port.Write(p)
}

// Command sends a SCPI/ASCII command to the serial port. The command can be
// optionally formatted according to a format specifier. An endmark character,
// such as newline, is automatically added to the end of the string.
func (d *Device) Command(ctx context.Context, cmd string, a ...any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.hwHandshaking {
		if err := d.napIfDataSetNotReady(ctx); err != nil {
			return err
		}
	}
	if len(a) > 0 {
		cmd = fmt.Sprintf(cmd, a...)
	}
	cmd = strings.TrimSpace(cmd) + string(d.endMark)
	if _, err := d.WriteBinary(ctx, []byte(cmd)); err != nil {
		return err
	}

	return sleepContext(ctx, d.delayTime)
}

// Query writes the given SCPI/ASCII command to the serial port and returns the
// response string. The device's endmark character (newline by default) is
// automatically added to the query command. The string returned is not
// stripped of any whitespace. The context is used for cancellation; if the
// context is canceled while waiting for a response, Query returns the context
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
		s, err := d.reader.ReadString(d.endMark)
		ch <- result{s, err}
	}()

	select {
	case <-ctx.Done():
		// Set a short read timeout to unblock the goroutine stuck on
		// ReadString, then wait for it to finish so we don't leak it or
		// race on the bufio.Reader.
		_ = d.port.SetReadTimeout(1 * time.Millisecond)
		<-ch
		_ = d.port.SetReadTimeout(d.readTimeout)
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
	timeout := time.NewTimer(d.readTimeout)
	defer timeout.Stop()
	ticker := time.NewTicker(d.delayTime)
	defer ticker.Stop()

	for {
		ready, err := isDSR(d.port)
		if err != nil {
			return err
		}
		if ready {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("%w after %s", ErrDSRNotReady, d.readTimeout)
		case <-ticker.C:
		}
	}
	// Sleep a bit longer once the Data Set Ready is true. Without this, the
	// Keysight E3631A DC power supply will sometimes hang when sending
	// commands/queries.
	return sleepContext(ctx, d.delayTime)
}

// sleepContext pauses for the given duration but returns early with the context
// error if the context is canceled.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
