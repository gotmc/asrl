# asrl

Go-based implementation of an Asynchronous Serial (ASRL) interface for
Interchangeable Virtual Instrument (IVI) drivers.

[![GoDoc][godoc badge]][godoc link]
[![Go Report Card][report badge]][report card]
[![License Badge][license badge]][LICENSE.txt]

## Overview

The [asrl][] package enables controlling test equipment (e.g., oscilloscopes,
function generators, multimeters, etc.) over serial port. While this package can
be used by itself to send Standard Commands for Programmable Instruments
([SCPI][]) commands to a piece of test equipment, it also serves to provide an
Instrument interface for both the [ivi][] and [visa][] packages. The [ivi][]
package provides standardized APIs for programming test instruments following
the [Interchangeable Virtual Instrument (IVI) standard][ivi-specs].

## Usage

```go
dev, err := asrl.NewDevice("ASRL::/dev/tty.usbserial-PX8X3YR6::9600::8N2::INSTR")
if err != nil {
    log.Fatal(err)
}
defer dev.Close()

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
```

## Documentation

Documentation can be found at <https://pkg.go.dev/github.com/gotmc/asrl>.

## Contributing

Contributions are welcome! To contribute please:

1. Fork the repository
2. Create a feature branch
3. Code
4. Submit a [pull request][]

### Development Dependencies

- [just][] - task runner that replaces [GNU Make][make]

### Testing

Prior to submitting a [pull request][], please run:

```bash
$ just check
$ just lint
```

To update and view the test coverage report:

```bash
$ just cover
```

## License

[asrl][] is released under the MIT license. Please see the [LICENSE.txt][] file
for more information.

[asrl]: https://github.com/gotmc/asrl
[godoc badge]: https://pkg.go.dev/badge/github.com/gotmc/asrl
[godoc link]: https://pkg.go.dev/github.com/gotmc/asrl
[ivi]: https://github.com/gotmc/ivi
[ivi-specs]: http://www.ivifoundation.org/specifications/
[just]: https://just.systems/
[LICENSE.txt]: https://github.com/gotmc/asrl/blob/master/LICENSE.txt
[license badge]: https://img.shields.io/badge/license-MIT-blue.svg
[make]: https://www.gnu.org/software/make/
[pull request]: https://help.github.com/articles/using-pull-requests
[report badge]: https://goreportcard.com/badge/github.com/gotmc/asrl
[report card]: https://goreportcard.com/report/github.com/gotmc/asrl
[scpi]: http://www.ivifoundation.org/scpi/
[visa]: https://github.com/gotmc/visa
