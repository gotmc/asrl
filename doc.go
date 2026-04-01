// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

// Package asrl provides an Asynchronous Serial (ASRL) interface for
// controlling test equipment via serial ports using SCPI commands. It
// implements the VISA ASRL resource string format and serves as an instrument
// driver for the ivi and visa packages.
//
// This package is part of the gotmc ecosystem. The visa package
// (github.com/gotmc/visa) defines a common interface for instrument
// communication across different transports (GPIB, USB, TCP/IP, serial). The
// asrl package provides the serial transport implementation. The ivi package
// (github.com/gotmc/ivi) builds on top of visa to provide standardized,
// instrument-class-specific APIs following the IVI Foundation specifications.
//
// Devices are addressed using VISA resource strings of the form:
//
//	ASRL::<port>::<baud>::<dataflow>::INSTR
//
// For example:
//
//	ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR
//
// Supported dataflow values are 8N1 (default), 8N2, 7E2, 7E1, and 7O1.
package asrl
