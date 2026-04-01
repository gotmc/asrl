// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"go.bug.st/serial"
)

// mockPort implements serial.Port for testing.
type mockPort struct {
	readBuf     *bytes.Buffer
	writeBuf    *bytes.Buffer
	readTimeout time.Duration
	closed      bool
	dsrReady    bool
	dsrErr      error
	readErr     error
	writeErr    error
	closeErr    error
}

func newMockPort(readData string) *mockPort {
	return &mockPort{
		readBuf:  bytes.NewBufferString(readData),
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockPort) Read(p []byte) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	return m.readBuf.Read(p)
}

func (m *mockPort) Write(p []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuf.Write(p)
}

func (m *mockPort) SetMode(_ *serial.Mode) error         { return nil }
func (m *mockPort) Drain() error                         { return nil }
func (m *mockPort) ResetInputBuffer() error              { return nil }
func (m *mockPort) ResetOutputBuffer() error             { return nil }
func (m *mockPort) SetDTR(_ bool) error                  { return nil }
func (m *mockPort) SetRTS(_ bool) error                  { return nil }
func (m *mockPort) Close() error                         { return m.closeErr }
func (m *mockPort) Break(_ time.Duration) error          { return nil }
func (m *mockPort) SetReadTimeout(t time.Duration) error { m.readTimeout = t; return nil }
func (m *mockPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	if m.dsrErr != nil {
		return nil, m.dsrErr
	}
	return &serial.ModemStatusBits{DSR: m.dsrReady}, nil
}

func newTestDevice(mp *mockPort) *Device {
	return &Device{
		port:        mp,
		reader:      bufio.NewReader(mp),
		EndMark:     '\n',
		DelayTime:   1 * time.Millisecond,
		ReadTimeout: 100 * time.Millisecond,
	}
}

func TestRead(t *testing.T) {
	t.Parallel()
	mp := newMockPort("hello")
	d := newTestDevice(mp)
	buf := make([]byte, 16)
	n, err := d.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(buf[:n]); got != "hello" {
		t.Errorf("Read = %q, want %q", got, "hello")
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	n, err := d.Write([]byte("world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}
	if got := mp.writeBuf.String(); got != "world" {
		t.Errorf("written = %q, want %q", got, "world")
	}
}

func TestWriteString(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	n, err := d.WriteString("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("WriteString returned %d, want 4", n)
	}
	if got := mp.writeBuf.String(); got != "test" {
		t.Errorf("written = %q, want %q", got, "test")
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	if err := d.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloseError(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	mp.closeErr = errors.New("close failed")
	d := newTestDevice(mp)
	if err := d.Close(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteContext(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx := context.Background()
	n, err := d.WriteContext(ctx, []byte("data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("WriteContext returned %d, want 4", n)
	}
	if got := mp.writeBuf.String(); got != "data" {
		t.Errorf("written = %q, want %q", got, "data")
	}
}

func TestWriteContextCanceled(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := d.WriteContext(ctx, []byte("data"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if mp.writeBuf.Len() != 0 {
		t.Error("expected no data written when context is canceled")
	}
}

func TestWriteStringContext(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx := context.Background()
	n, err := d.WriteStringContext(ctx, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteStringContext returned %d, want 5", n)
	}
}

func TestReadContext(t *testing.T) {
	t.Parallel()
	mp := newMockPort("response")
	d := newTestDevice(mp)
	ctx := context.Background()
	buf := make([]byte, 16)
	n, err := d.ReadContext(ctx, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(buf[:n]); got != "response" {
		t.Errorf("ReadContext = %q, want %q", got, "response")
	}
}

func TestReadContextCanceled(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := d.ReadContext(ctx, make([]byte, 16))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx := context.Background()
	if err := d.Command(ctx, "FREQ %d", 100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mp.writeBuf.String(); got != "FREQ 100\n" {
		t.Errorf("written = %q, want %q", got, "FREQ 100\n")
	}
}

func TestCommandNoArgs(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx := context.Background()
	if err := d.Command(ctx, "*RST"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mp.writeBuf.String(); got != "*RST\n" {
		t.Errorf("written = %q, want %q", got, "*RST\n")
	}
}

func TestCommandCanceled(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := d.Command(ctx, "*RST")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestCommandTrimsWhitespace(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx := context.Background()
	if err := d.Command(ctx, "  *RST  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mp.writeBuf.String(); got != "*RST\n" {
		t.Errorf("written = %q, want %q", got, "*RST\n")
	}
}

func TestCommandWriteError(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	mp.writeErr = errors.New("write failed")
	d := newTestDevice(mp)
	ctx := context.Background()
	if err := d.Command(ctx, "*RST"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()
	mp := newMockPort("Stanford Research Systems,DS345\n")
	d := newTestDevice(mp)
	ctx := context.Background()
	got, err := d.Query(ctx, "*IDN?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Stanford Research Systems,DS345\n"
	if got != want {
		t.Errorf("Query = %q, want %q", got, want)
	}
	if written := mp.writeBuf.String(); written != "*IDN?\n" {
		t.Errorf("written = %q, want %q", written, "*IDN?\n")
	}
}

func TestQueryCanceled(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := d.Query(ctx, "*IDN?")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestCommandWithHWHandshaking(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	mp.dsrReady = true
	d := newTestDevice(mp)
	d.HWHandshaking = true
	ctx := context.Background()
	if err := d.Command(ctx, "*RST"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mp.writeBuf.String(); got != "*RST\n" {
		t.Errorf("written = %q, want %q", got, "*RST\n")
	}
}

func TestCommandHWHandshakingDSRError(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	mp.dsrErr = errors.New("modem error")
	d := newTestDevice(mp)
	d.HWHandshaking = true
	ctx := context.Background()
	err := d.Command(ctx, "*RST")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCommandHWHandshakingTimeout(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	mp.dsrReady = false
	d := newTestDevice(mp)
	d.HWHandshaking = true
	d.ReadTimeout = 5 * time.Millisecond
	d.DelayTime = 1 * time.Millisecond
	ctx := context.Background()
	err := d.Command(ctx, "*RST")
	if err == nil {
		t.Fatal("expected DSR timeout error, got nil")
	}
}

func TestIsDSR(t *testing.T) {
	t.Parallel()

	t.Run("ready", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		mp.dsrReady = true
		ready, err := isDSR(mp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ready {
			t.Error("isDSR = false, want true")
		}
	})

	t.Run("not ready", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		mp.dsrReady = false
		ready, err := isDSR(mp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ready {
			t.Error("isDSR = true, want false")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		mp.dsrErr = errors.New("modem error")
		_, err := isDSR(mp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDeviceOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithEndMark", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		d := newTestDevice(mp)
		WithEndMark('\r')(d)
		if d.EndMark != '\r' {
			t.Errorf("EndMark = %q, want %q", d.EndMark, '\r')
		}
	})

	t.Run("WithHWHandshaking", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		d := newTestDevice(mp)
		WithHWHandshaking(true)(d)
		if !d.HWHandshaking {
			t.Error("HWHandshaking = false, want true")
		}
	})

	t.Run("WithDelayTime", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		d := newTestDevice(mp)
		WithDelayTime(100 * time.Millisecond)(d)
		if d.DelayTime != 100*time.Millisecond {
			t.Errorf("DelayTime = %v, want %v", d.DelayTime, 100*time.Millisecond)
		}
	})

	t.Run("WithReadTimeout", func(t *testing.T) {
		t.Parallel()
		mp := newMockPort("")
		d := newTestDevice(mp)
		WithReadTimeout(10 * time.Second)(d)
		if d.ReadTimeout != 10*time.Second {
			t.Errorf("ReadTimeout = %v, want %v", d.ReadTimeout, 10*time.Second)
		}
	})
}

func TestCommandWithCustomEndMark(t *testing.T) {
	t.Parallel()
	mp := newMockPort("")
	d := newTestDevice(mp)
	d.EndMark = '\r'
	ctx := context.Background()
	if err := d.Command(ctx, "*RST"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mp.writeBuf.String(); got != "*RST\r" {
		t.Errorf("written = %q, want %q", got, "*RST\r")
	}
}
