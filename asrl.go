// Copyright (c) 2017-2024 The asrl developers. All rights reserved.
// Project site: https://github.com/gotmc/asrl
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package asrl

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Device models a serial device and implements the ivi.Driver interface.
type Device struct {
	EndMark   byte
	DelayTime time.Duration
	port      serial.Port
}

// NewDevice opens a serial Device using the given VISA address resource string.
func NewDevice(address string) (*Device, error) {
	v, err := NewVisaResource(address)
	if err != nil {
		return nil, err
	}
	log.Printf("Baud rate = %d", v.baud)
	log.Printf("Parity = %v", v.parity)
	log.Printf("Data bits = %d", v.dataBits)
	log.Printf("Stop bits = %v", v.stopBits)

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

	return &Device{
		port:      port,
		EndMark:   '\n',
		DelayTime: 70 * time.Millisecond,
	}, nil
}

// Write writes the given data to the network connection.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.port.Write(p)
}

// Read reads from the network connection into the given byte slice.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.port.Read(p)
}

// Close closes the underlying network connection.
func (d *Device) Close() error {
	return d.port.Close()
}

// WriteString writes a string using the underlying network connection.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends the SCPI/ASCII command to the underlying network connection. A
// newline character is automatically added to the end of the string.
func (d *Device) Command(format string, a ...any) error {
	d.napIfDataSetNotReady()
	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	cmd = strings.TrimSpace(cmd) + string(d.EndMark)
	_, err := d.WriteString(cmd)
	if err != nil {
		return err
	}

	return err
}

// Query writes the given string to the underlying network connection and
// returns a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(cmd string) (string, error) {
	log.Printf("starting query for %s", cmd)
	err := d.Command(cmd)
	log.Printf("command sent to query for %s", cmd)
	if err != nil {
		log.Printf("error received from command sent to query for %s", cmd)
		return "", err
	}
	log.Printf("Just before reading bufio string for query for %s", cmd)
	return bufio.NewReader(d.port).ReadString('\n')
}

func isDSR(port serial.Port) bool {
	msb, err := port.GetModemStatusBits()
	if err != nil {
		log.Printf("error getting modem status bits: %s", err)
	}
	return msb.DSR
}

func (d *Device) napIfDataSetNotReady() {
	//------------------------------------------------------------------------//
	//
	//                       ___====-_  _-====___
	//                 _--^^^#####//      \\#####^^^--_
	//              _-^##########// (    ) \\##########^-_
	//             -############//  |\^^/|  \\############-
	//           _/############//   (@::@)   \\############\_
	//          /#############((     \\//     ))#############\
	//         -###############\\    (oo)    //###############-
	//        -#################\\  / VV \  //#################-
	//       -###################\\/      \//###################-
	//      _#/|##########/\######(   /\   )######/\##########|\#_
	//      |/ |#/\#/\#/\/  \#/\##\  |  |  /##/\#/  \/\#/\#/\#| \|
	//      `  |/  V  V  `   V  \#\| |  | |/#/  V   '  V  V  \|  '
	//         `   `  `      `   / | |  | | \   '      '  '   '
	//                          (  | |  | |  )
	//                         __\ | |  | | /__
	//                        (vvv(VVV)(VVV)vvv)
	//------------------------------------------------------------------------//
	// If I use 40 ms instead of 50 ms for the delay time, the Keysight E3631A DC
	// power supply will hang when sending commands/queries. Using 50 ms causes
	// the power supply to hang sometimes. I'm currently using 70 ms to be safe.
	//------------------------------------------------------------------------//
	for !isDSR(d.port) {
		log.Printf("DSR is false, so napping for %s", d.DelayTime)
		time.Sleep(d.DelayTime)
	}
	//------------------------------------------------------------------------//
	// Sleep a bit longer once the Data Set Ready is true. Without this, the
	// Keysight E3631A DC power supply will sometimes hang when sending
	// commands/queries.
	time.Sleep(d.DelayTime)
	//------------------------------------------------------------------------//
}
