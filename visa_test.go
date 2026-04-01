// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
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
	t.Parallel()

	testCases := []struct {
		name           string
		resourceString string
		interfaceType  string
		address        string
		baud           int
		dataBits       int
		parity         serial.Parity
		stopBits       serial.StopBits
		resourceClass  string
		wantErr        error
	}{
		{
			name:           "9600 baud 8N2",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N2::INSTR",
			interfaceType:  "ASRL",
			address:        "/dev/tty.usbserial-PX484GRU",
			baud:           9600,
			dataBits:       8,
			parity:         serial.NoParity,
			stopBits:       serial.TwoStopBits,
			resourceClass:  "INSTR",
		},
		{
			name:           "115200 baud 8N1",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::115200::8N1::INSTR",
			interfaceType:  "ASRL",
			address:        "/dev/tty.usbserial-PX484GRU",
			baud:           115200,
			dataBits:       8,
			parity:         serial.NoParity,
			stopBits:       serial.OneStopBit,
			resourceClass:  "INSTR",
		},
		{
			name:           "115200 baud 7E2",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::115200::7E2::INSTR",
			interfaceType:  "ASRL",
			address:        "/dev/tty.usbserial-PX484GRU",
			baud:           115200,
			dataBits:       7,
			parity:         serial.EvenParity,
			stopBits:       serial.TwoStopBits,
			resourceClass:  "INSTR",
		},
		{
			name:           "115200 baud 7E1",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::115200::7E1::INSTR",
			interfaceType:  "ASRL",
			address:        "/dev/tty.usbserial-PX484GRU",
			baud:           115200,
			dataBits:       7,
			parity:         serial.EvenParity,
			stopBits:       serial.OneStopBit,
			resourceClass:  "INSTR",
		},
		{
			name:           "115200 baud 7O1",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::115200::7O1::INSTR",
			interfaceType:  "ASRL",
			address:        "/dev/tty.usbserial-PX484GRU",
			baud:           115200,
			dataBits:       7,
			parity:         serial.OddParity,
			stopBits:       serial.OneStopBit,
			resourceClass:  "INSTR",
		},
		{
			name:           "completely invalid string",
			resourceString: "not-a-visa-string",
			wantErr:        ErrInvalidResource,
		},
		{
			name:           "unsupported dataflow",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::9600::9N1::INSTR",
			wantErr:        ErrUnsupportedDataflow,
		},
		{
			name:           "empty string",
			resourceString: "",
			wantErr:        ErrInvalidResource,
		},
		{
			name:           "missing INSTR resource class",
			resourceString: "ASRL::/dev/tty.usbserial-PX484GRU::9600::8N1::OTHER",
			wantErr:        ErrInvalidResource,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resource, err := NewVisaResource(tc.resourceString)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resource.interfaceType != tc.interfaceType {
				t.Errorf("interfaceType = %s, want %s", resource.interfaceType, tc.interfaceType)
			}
			if resource.address != tc.address {
				t.Errorf("address = %s, want %s", resource.address, tc.address)
			}
			if resource.baud != tc.baud {
				t.Errorf("baud = %d, want %d", resource.baud, tc.baud)
			}
			if resource.dataBits != tc.dataBits {
				t.Errorf("dataBits = %d, want %d", resource.dataBits, tc.dataBits)
			}
			if resource.parity != tc.parity {
				t.Errorf("parity = %d, want %d", resource.parity, tc.parity)
			}
			if resource.stopBits != tc.stopBits {
				t.Errorf("stopBits = %d, want %d", resource.stopBits, tc.stopBits)
			}
			if resource.resourceClass != tc.resourceClass {
				t.Errorf("resourceClass = %s, want %s", resource.resourceClass, tc.resourceClass)
			}
		})
	}
}
