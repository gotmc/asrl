// Copyright (c) 2017-2021 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"errors"
	"testing"

	"go.bug.st/serial"
)

func TestParsingVisaResourceString(t *testing.T) {
	testCases := []struct {
		resourceString string
		interfaceType  string
		address        string
		baud           int
		dataBits       int
		parity         serial.Parity
		stopBits       serial.StopBits
		resourceClass  string
		isError        bool
		errorString    error
	}{
		{
			"ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR",
			"ASRL", "/dev/tty.usbserial-PX484GRU",
			9600, 8, serial.NoParity, serial.TwoStopBits,
			"INSTR", false, errors.New(""),
		},
		{
			"ASRL::/dev/tty.usbserial-PX484GRU::115200::8N1::INSTR",
			"ASRL", "/dev/tty.usbserial-PX484GRU",
			115200, 8, serial.NoParity, serial.OneStopBit,
			"INSTR", false, errors.New(""),
		},
		{
			"ASRL::/dev/tty.usbserial-PX484GRU::115200::7E2::INSTR",
			"ASRL", "/dev/tty.usbserial-PX484GRU",
			115200, 7, serial.EvenParity, serial.TwoStopBits,
			"INSTR", false, errors.New(""),
		},
		{
			"ASRL::/dev/tty.usbserial-PX484GRU::115200::7E1::INSTR",
			"ASRL", "/dev/tty.usbserial-PX484GRU",
			115200, 7, serial.EvenParity, serial.OneStopBit,
			"INSTR", false, errors.New(""),
		},
		{
			"ASRL::/dev/tty.usbserial-PX484GRU::115200::7O1::INSTR",
			"ASRL", "/dev/tty.usbserial-PX484GRU",
			115200, 7, serial.OddParity, serial.OneStopBit,
			"INSTR", false, errors.New(""),
		},
	}
	for _, testCase := range testCases {
		resource, err := NewVisaResource(testCase.resourceString)
		if resource.interfaceType != testCase.interfaceType {
			t.Errorf(
				"interfaceType == %s, want %s for resource %s",
				resource.interfaceType,
				testCase.interfaceType,
				testCase.resourceString,
			)
		}
		if resource.address != testCase.address {
			t.Errorf(
				"address == %s, want %s for resource %s",
				resource.address,
				testCase.address,
				testCase.resourceString,
			)
		}
		if resource.baud != testCase.baud {
			t.Errorf(
				"baud == %d, want %d for resource %s",
				resource.baud,
				testCase.baud,
				testCase.resourceString,
			)
		}
		if resource.dataBits != testCase.dataBits {
			t.Errorf(
				"dataBits == %d, want %d for resource %s",
				resource.dataBits,
				testCase.dataBits,
				testCase.resourceString,
			)
		}
		if resource.parity != testCase.parity {
			t.Errorf(
				"parity == %d, want %d for resource %s",
				resource.parity,
				testCase.parity,
				testCase.resourceString,
			)
		}
		if resource.stopBits != testCase.stopBits {
			t.Errorf(
				"stopBits == %d, want %d for resource %s",
				resource.stopBits,
				testCase.stopBits,
				testCase.resourceString,
			)
		}
		if resource.resourceClass != testCase.resourceClass {
			t.Errorf(
				"resourceClass == %s, want %s for resource %s",
				resource.resourceClass,
				testCase.resourceClass,
				testCase.resourceString,
			)
		}
		if err != nil && testCase.isError {
			if err.Error() != testCase.errorString.Error() {
				t.Errorf(
					"err == %s, want %s for resource %s",
					err,
					testCase.errorString,
					testCase.resourceString,
				)
			}
		}
		if err != nil && !testCase.isError {
			t.Errorf("Unhandled error: %q for resource %s", err, testCase.resourceString)
		}
	}
}
