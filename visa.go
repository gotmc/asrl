// Copyright (c) 2017-2024 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"errors"
	"regexp"
	"strconv"

	"go.bug.st/serial"
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
func NewVisaResource(resourceString string) (visa *VisaResource, err error) {
	visa = &VisaResource{
		resourceString: resourceString,
	}
	regString := `^(?P<interfaceType>ASRL)(?P<boardIndex>\d*)::` +
		`(?P<address>[^\s]+)::` +
		`(?P<baud>\d+)::` +
		`(?P<dataflow>\d{1}\w{1}\d{1})::` +
		`(?P<resourceClass>INSTR)$`

	re := regexp.MustCompile(regString)
	res := re.FindStringSubmatch(resourceString)
	subexpNames := re.SubexpNames()
	matchMap := map[string]string{}
	for i, n := range res {
		matchMap[subexpNames[i]] = string(n)
	}

	if matchMap["interfaceType"] != "ASRL" {
		return visa, errors.New("visa: interface type was not ASRL")
	}
	visa.interfaceType = "ASRL"

	if matchMap["resourceClass"] != "INSTR" {
		return visa, errors.New("visa: resource class was not INSTR")
	}
	visa.resourceClass = "INSTR"

	if matchMap["address"] != "" {
		visa.address = matchMap["address"]
	}

	if matchMap["baud"] != "" {
		baud, err := strconv.ParseUint(matchMap["baud"], 10, 64)
		if err != nil {
			return visa, errors.New("visa: baud error")
		}
		visa.baud = int(baud)
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
			return visa, errors.New("visa: dataflow error")
		}
	} else {
		visa.dataBits = 8
		visa.parity = serial.NoParity
		visa.stopBits = serial.OneStopBit
	}

	return visa, nil
}
