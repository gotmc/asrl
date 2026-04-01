// Copyright (c) 2017-2026 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"go.bug.st/serial"
)

// Sentinel errors returned by NewVisaResource.
var (
	ErrInvalidResource     = errors.New("visa: invalid VISA resource string")
	ErrInvalidInterfaceType = errors.New("visa: interface type was not ASRL")
	ErrInvalidResourceClass = errors.New("visa: resource class was not INSTR")
	ErrInvalidBaud         = errors.New("visa: invalid baud")
	ErrUnsupportedDataflow = errors.New("visa: unsupported dataflow")
)

var visaResourceRE = regexp.MustCompile(
	`^(?P<interfaceType>ASRL)(?P<boardIndex>\d*)::` +
		`(?P<address>[^\s]+)::` +
		`(?P<baud>\d+)::` +
		`(?P<dataflow>\d{1}\w{1}\d{1})::` +
		`(?P<resourceClass>INSTR)$`,
)

// VisaResource represents a VISA enabled piece of test equipment.
type VisaResource struct {
	resourceString string
	interfaceType  string
	address        string
	baud           int
	dataBits       int
	parity         serial.Parity
	stopBits       serial.StopBits
	resourceClass  string
}

// NewVisaResource creates a new VisaResource using the given VISA
// resourceString. If the dataflow isn't provided as part of the VISA resource
// string, the dataflow will default to 8N1.
func NewVisaResource(resourceString string) (*VisaResource, error) {
	res := visaResourceRE.FindStringSubmatch(resourceString)
	if res == nil {
		return nil, ErrInvalidResource
	}
	subexpNames := visaResourceRE.SubexpNames()
	matchMap := map[string]string{}
	for i, n := range res {
		matchMap[subexpNames[i]] = n
	}

	if matchMap["interfaceType"] != "ASRL" {
		return nil, ErrInvalidInterfaceType
	}

	if matchMap["resourceClass"] != "INSTR" {
		return nil, ErrInvalidResourceClass
	}

	visa := &VisaResource{
		resourceString: resourceString,
		interfaceType:  "ASRL",
		resourceClass:  "INSTR",
		address:        matchMap["address"],
	}

	if matchMap["baud"] != "" {
		baud, err := strconv.Atoi(matchMap["baud"])
		if err != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrInvalidBaud, matchMap["baud"], err)
		}
		visa.baud = baud
	}

	if matchMap["dataflow"] != "" {
		switch matchMap["dataflow"] {
		case "8N1":
			visa.dataBits = 8
			visa.parity = serial.NoParity
			visa.stopBits = serial.OneStopBit
		case "8N2":
			visa.dataBits = 8
			visa.parity = serial.NoParity
			visa.stopBits = serial.TwoStopBits
		case "7E2":
			visa.dataBits = 7
			visa.parity = serial.EvenParity
			visa.stopBits = serial.TwoStopBits
		case "7E1":
			visa.dataBits = 7
			visa.parity = serial.EvenParity
			visa.stopBits = serial.OneStopBit
		case "7O1":
			visa.dataBits = 7
			visa.parity = serial.OddParity
			visa.stopBits = serial.OneStopBit
		default:
			return nil, fmt.Errorf("%w %q", ErrUnsupportedDataflow, matchMap["dataflow"])
		}
	} else {
		visa.dataBits = 8
		visa.parity = serial.NoParity
		visa.stopBits = serial.OneStopBit
	}

	return visa, nil
}

// String returns the original VISA resource string.
func (v *VisaResource) String() string {
	return v.resourceString
}

// InterfaceType returns the VISA interface type (e.g., "ASRL").
func (v *VisaResource) InterfaceType() string {
	return v.interfaceType
}

// Address returns the serial port address.
func (v *VisaResource) Address() string {
	return v.address
}

// Baud returns the baud rate.
func (v *VisaResource) Baud() int {
	return v.baud
}

// DataBits returns the number of data bits.
func (v *VisaResource) DataBits() int {
	return v.dataBits
}

// Parity returns the parity setting.
func (v *VisaResource) Parity() serial.Parity {
	return v.parity
}

// StopBits returns the stop bits setting.
func (v *VisaResource) StopBits() serial.StopBits {
	return v.stopBits
}

// ResourceClass returns the VISA resource class (e.g., "INSTR").
func (v *VisaResource) ResourceClass() string {
	return v.resourceClass
}
