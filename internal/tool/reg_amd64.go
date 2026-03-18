// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package tool

import (
	"fmt"

	"github.com/usbarmory/tamago/amd64"

	"github.com/usbarmory/kanzashi/internal/reg"
)

var fault = -1

func checkFault() (err error) {
	if amd64.Faulty() {
		err = fmt.Errorf("fault")
	}
	return
}

func Read32(addr uint32) (val uint32, err error) {
	fmt.Printf("[kanzashi]  read32 %#08x", addr)
	val = reg.Read(addr)
	err = checkFault()
	fmt.Printf(" => %#08x (%v)\n", val, err)
	return
}

func Write32(addr uint32, val uint32) (err error) {
	fmt.Printf("[kanzashi] write32 %#08x <= %#08x", addr, val)
	reg.Write32(addr, val)
	err = checkFault()
	fmt.Printf(" (%v)\n", err)
	return
}

func Read64(addr uint64) (val uint64, err error) {
	fmt.Printf("[kanzashi]  read64 %#016x", addr)
	val = reg.Read64(addr)
	err = checkFault()
	fmt.Printf(" => %#016x (%v)\n", val, err)
	return
}

func Write64(addr uint64, val uint64) (err error) {
	fmt.Printf("[kanzashi] write64 %#016x <= %#016x", addr, val)
	reg.Write64(addr, val)
	err = checkFault()
	fmt.Printf(" (%v)\n", err)
	return
}

func ReadMSR(addr uint64) (val uint64, err error) {
	fmt.Printf("[kanzashi]   rdmsr %#016x", addr)
	val = reg.ReadMSR(addr)
	err = checkFault()
	fmt.Printf(" => %#016x (%v)\n", val, err)
	return
}

func WriteMSR(addr uint64, val uint64) (err error) {
	fmt.Printf("[kanzashi]   wrmsr %#016x <= %#016x", addr, val)
	reg.WriteMSR(addr, val)
	err = checkFault()
	fmt.Printf(" (%v)\n", err)
	return checkFault()
}
