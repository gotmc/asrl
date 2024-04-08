# asrl

Go-based implementation of an Asynchronous Serial (ASRL) interface for IVI.

[![GoDoc][godoc badge]][godoc link]
[![Go Report Card][report badge]][report card]
[![License Badge][license badge]][LICENSE.txt]

## Overview

The [asrl][] package enables controlling test equipment (e.g., oscilloscopes,
function generators, multimeters, etc.) over serial port. While this package can
be used by itself to send [SCPI][] commands to a piece of test equipment, it
also serves to provide an Instrument interface for both the [ivi][] and [visa][]
packages. The [ivi][] package provides standardized APIs for programming test
instruments following the [Interchangeable Virtual Instrument (IVI)
standard][ivi-specs].

## Implementations

### Prologix

The `prologix` package provides a serial interface for the Prologix GPIB-USB
Controller using the Virtual COM Port (VCP) driver. To work, you must download
the drivers for FT245R chip from FTDI website (www.ftdichip.com).

## Installation

```bash
$ go get github.com/gotmc/asrl
```

## Documentation

Documentation can be found at either:

- <https://godoc.org/github.com/gotmc/asrl>
- <http://localhost:6060/pkg/github.com/gotmc/asrl/> after running `$
godoc -http=:6060`

## Contributing

Contributions are welcome! To contribute please:

1. Fork the repository
2. Create a feature branch
3. Code
4. Submit a [pull request][]

### Testing

Prior to submitting a [pull request][], please run the tests using either [GNU
Make][make]:

```bash
$ make check
$ make lint
```

or you can use [Just][]:

```bash
$ just check
$ just lint
```

To update and view the test coverage report using [Make][] run:

```bash
$ make cover
```

or you can use [Just][]:

```bash
$ just cover
```

## License

[asrl][] is released under the MIT license. Please see the [LICENSE.txt][] file
for more information.

[asrl]: https://github.com/gotmc/asrl
[godoc badge]: https://godoc.org/github.com/gotmc/asrl?status.svg
[godoc link]: https://godoc.org/github.com/gotmc/asrl
[ivi]: https://github.com/gotmc/ivi
[ivi-foundation]: http://www.ivifoundation.org/
[ivi-specs]: http://www.ivifoundation.org/specifications/
[just]: https://just.systems/man/en/
[LICENSE.txt]: https://github.com/gotmc/lxi/blob/master/LICENSE.txt
[license badge]: https://img.shields.io/badge/license-MIT-blue.svg
[make]: https://www.gnu.org/software/make/
[pull request]: https://help.github.com/articles/using-pull-requests
[report badge]: https://goreportcard.com/badge/github.com/gotmc/asrl
[report card]: https://goreportcard.com/report/github.com/gotmc/asrl
[scpi]: http://www.ivifoundation.org/scpi/
[visa]: https://github.com/gotmc/visa
