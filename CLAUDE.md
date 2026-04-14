# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`asrl` is a Go package providing an Asynchronous Serial (ASRL) interface for controlling test equipment (oscilloscopes, function generators, power supplies, etc.) over serial ports. It implements VISA (Virtual Instrument Software Architecture) resource string parsing and SCPI command communication. Part of the [gotmc](https://github.com/gotmc) ecosystem alongside the `ivi` and `visa` packages.

## Build & Test Commands

```bash
just check          # Format (go fmt) and vet (go vet)
just unit           # Run unit tests with race detection (-short flag); runs check first
just lint           # Lint with golangci-lint (uses .golangci.yaml config)
just cover          # Generate and open HTML coverage report
just tidy           # go mod tidy and verify
just updateall      # Update all dependencies
```

Run a single test:
```bash
go test -run TestParsingVisaResourceString -v ./...
```

Run examples (require physical hardware and serial port):
```bash
just e3631a /dev/tty.usbserial-PX8X3YR6
just ds345 /dev/tty.usbserial-XXXX
```

## Architecture

- **`asrl.go`** — `Device` struct wrapping `go.bug.st/serial.Port`. Implements `io.Reader`, `io.Writer`, and `io.Closer`. Provides `Command()` for sending SCPI commands and `Query()` for send-and-read. Binary variants (`ReadBinary`, `WriteBinary`) provide context-aware raw I/O without terminator handling. Handles hardware handshaking via DSR (Data Set Ready) polling with configurable `DelayTime`.
- **`visa.go`** — `VisaResource` struct and parser. Parses VISA address strings of the form `ASRL::<port>::<baud>::<dataflow>::INSTR` (e.g., `ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR`). Supported dataflow values: `8N1`, `8N2`, `7E2`, `7E1`, `7O1`. Defaults to `8N1`.
- **`examples/`** — Standalone example applications for specific instruments (Keysight E3631A power supply, SRS DS345 function generator).

## Key Details

- **Entry point:** `NewDevice(ctx, address, opts...)` parses a VISA resource string via `NewVisaResource()`, then opens the serial port with the parsed settings. Functional options (`WithEndMark`, `WithHWHandshaking`, `WithDelayTime`, `WithReadTimeout`) configure the device at construction time.
- `Command()` and `Query()` accept `context.Context` for cancellation support. `Command()` auto-appends the EndMark character and uses `WriteBinary` internally. `Query()` sends a command then reads the response. All delays use `sleepContext` which returns early on context cancellation.
- `Query()` spawns a goroutine for the blocking `ReadString` call. On context cancellation, it sets a short read timeout to unblock the goroutine, waits for it to finish, then resets the reader — this prevents goroutine leaks and races on the `bufio.Reader`. `ReadBinary` uses the same pattern for raw reads.
- `WriteBinary` checks context before writing but does not use a goroutine — serial writes are typically non-blocking, so `SetReadTimeout` cannot unblock them.
- Hardware handshaking (`HWHandshaking`) is disabled by default. When enabled, `Command()` polls DSR via `time.Ticker`/`time.Timer` before writing, with a timeout based on `ReadTimeout`.
- **Sentinel errors:** `ErrInvalidResource`, `ErrInvalidInterfaceType`, `ErrInvalidResourceClass`, `ErrInvalidBaud`, `ErrUnsupportedDataflow` (in `visa.go`) and `ErrDSRNotReady` (in `asrl.go`). All support `errors.Is()`.
- `DelayTime` (default 70ms) is critical for reliable communication — values below ~50ms can cause hangs with certain instruments (e.g., Keysight E3631A).
- `ReadTimeout` (default 5s) is set on the serial port and also used as the DSR polling deadline.
- `Close()` uses `time.Sleep` (not context-aware) since it has no context parameter.
- Only dependency: `go.bug.st/serial` for serial port access.
- Requires Go 1.25+.

## Testing

- **`visa_test.go`** — Table-driven tests for VISA resource string parsing, covering valid dataflow variants and error cases.
- **`asrl_test.go`** — Unit tests for `Device` methods using a `mockPort` that implements `serial.Port`. Covers `Read`, `Write`, `Close`, context-aware methods, `Command`, `Query`, DSR polling, and functional options. Tests use minimal `DelayTime` (1ms) and `ReadTimeout` (100ms) to run fast.
