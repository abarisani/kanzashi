// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"log"

	"github.com/usbarmory/go-boot/shell"

	"github.com/usbarmory/kanzashi/cmd"
	"github.com/usbarmory/kanzashi/internal/platform"
	"github.com/usbarmory/kanzashi/internal/tool"
)

//go:embed log.txt
var pastSession string

func init() {
	log.SetFlags(0)
	log.Printf("\n")
}

func main() {
	// The `tamago.patch` included in this repository modifies tamago
	// exception handler to ignore faults and skip the 2-bytes instruction
	// which caused it, the fault is reported in a boolean for tooling
	// awareness of the event.
	//
	// This check ensures the patch is applied.
	log.Printf("[kanzashi] checking exception handler patch")
	if _, err := tool.Read32(0); err == nil {
		log.Fatalf("[kanxashi] exception handler override error")
	}

	log.Printf("[kanzashi] starting network")
	if err := platform.StartNetwork(); err != nil {
		log.Fatal(err)
	}

	cmd.PastSession = pastSession

	console := &shell.Interface{
		Banner:     cmd.Banner,
		ReadWriter: platform.Terminal,
	}

	// start interactive shell
	console.Start(true)

	log.Printf("[kanzashi] graceful exit")
}
