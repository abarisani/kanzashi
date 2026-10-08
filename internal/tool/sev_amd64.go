// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package tool

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/kvm/sev"
)

var GHCB *sev.GHCB

func VMExit(req []byte, code, info1, info2, scratch uint64) (res []byte, err error) {
	if GHCB == nil {
		return nil, errors.New("GHCB instance not set")
	}

	fmt.Printf("[kanzashi] vmgexit %#016x %#016x  %#016x %#016x ", code, info1, info2, scratch)

	addr := GHCB.Layout.Start()
	GHCB.Layout.Write(addr, 0, req)

	err = GHCB.Exit(code, info1, info2, scratch)
	fmt.Printf(" (%v)\n", err)

	res = make([]byte, 4096)
	GHCB.Layout.Read(addr, len(res), res)

	return
}
