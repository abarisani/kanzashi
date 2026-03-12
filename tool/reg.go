package tool

import (
	"fmt"

	"github.com/usbarmory/tamago/amd64"

	"github.com/usbarmory/kanzashi/reg"
)

var fault = -1

func checkFault() (err error) {
	if amd64.Faulty() {
		err = fmt.Errorf("fault")
	}
	return
}

func Read32(addr uint32) (val uint32, err error) {
	val = reg.Read(addr)
	return val, checkFault()
}

func Write32(addr uint32, val uint32) (err error) {
	reg.Write32(addr, val)
	return checkFault()
}

func Read64(addr uint64) (val uint64, err error) {
	val = reg.Read64(addr)
	return val, checkFault()
}

func Write64(addr uint64, val uint64) (err error) {
	reg.Write64(addr, val)
	return checkFault()
}

func ReadMSR(addr uint64) (val uint64, err error) {
	val = reg.ReadMSR(addr)
	return val, checkFault()
}

func WriteMSR(addr uint64, val uint64) (err error) {
	reg.WriteMSR(addr, val)
	return checkFault()
}
