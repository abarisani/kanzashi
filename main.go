// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/amd64"

	"github.com/usbarmory/kanzashi/internal/llm"
	"github.com/usbarmory/kanzashi/internal/network"
	"github.com/usbarmory/kanzashi/internal/tool"
)

//go:embed log.txt
var pastSession string

func init() {
	amd64.SystemExceptionHandler = nil
}

func main() {
	log.SetFlags(0)
	log.Printf("\n[kanzashi] starting network")

	// The `tamago.patch` included in this repository modifies tamago
	// exception handler to ignore faults and skip the 2-bytes instruction
	// which caused it, the fault is reported in a boolean for tooling
	// awareness of the event.
	//
	// This check ensures the patch is applied.
	if _, err := tool.Read32(0); err == nil {
		log.Fatalf("[kanxashi] exception handler override error")
	}

	if err := network.Start(); err != nil {
		log.Fatal(err)
	}

	system := `You are a hypervisor security researcher running bare metal inside a QEMU VM on AMD64.
You have direct access to physical memory via reg_read32, reg_write32, reg_read64, reg_write64, msr_read, msr_write tools.
Your goal is to autonomously explore the I/O peripheral range to find hypervisor escape attack surfaces and crash the hypervisor (not yourself).

Approach:
1. Assume a QEMU VM (either q35 or microvm).
2. Focus on PCI handling.
3. Do not document anomalous hypervisor responses, just aim for a privilege escalation on the hypervisor (QEMU).
4. Avoid touching the VirtIO network device that is providing access to your session.
5. If you crash the hypervisor it is a result, if you are still alive it means you had no effect on the host.

Think step by step and use the tools iteratively.`

	user := fmt.Sprintf("Begin autonomous security analysis of QEMU VM. Explore freely. Briefly explain what you are doing as you go.")

	if len(pastSession) > 0 {
		log.Printf("[kanzashi] using past session log (%d bytes)", len(pastSession))
		user += fmt.Sprintf("Here are the logs of the past sessions, resume from them:%s", pastSession)
	}

	log.Printf("[kanzashi] starting agentic QEMU audit...")
	log.Printf("\n%s\n%s\n\n", system, user)

	llm.RunAgent(context.Background(), system, user)

	log.Printf("[kanzashi] graceful exit")
}
