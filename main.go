package main

import (
	"context"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago/kvm/clock"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/usbarmory/kanzashi/reg"
)

const ClaudeApiKey = "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"

func init() {
	print("\n")
	microvm.AMD64.SetTime(kvmclock.Now().UnixNano())

	amd64.SystemExceptionHandler = nil
}

var fault = -1

func Read(addr uint32) (val uint32, err error) {
	val = reg.Read(addr)
	if amd64.Faulty() {
		err = fmt.Errorf("fault")
	}

	return
}

func Write(addr uint32, val uint32) (err error) {
	reg.Write32(addr, val)
	if amd64.Faulty() {
		err = fmt.Errorf("fault")
	}

	return
}

func main() {
	log.SetFlags(0)
	log.Printf("[kanzashi] starting network")

	if _, err := Read(0); err == nil {
		log.Fatalf("[kanxashi] exception handler override error")
	}

	if err := StartNetwork(); err != nil {
		log.Fatal(err)
	}

	log.Printf("[kanzashi] instantiating claude client")

	client := anthropic.NewClient(option.WithAPIKey(ClaudeApiKey))

	system := `You are a hypervisor security researcher running bare metal inside a QEMU microvm on AMD64.
You have direct access to physical memory via reg_read and reg_write tools.
Your goal is to autonomously explore the I/O peripheral range to find hypervisor escape attack surfaces and crash the hypervisor (not yourself).

Approach:
1. Assume a microvm QEMU machine.
2. Enumerate peripherals.
4. Do not document anomalous hypervisor responses, just aim for a DoS or privilege escalation on the hypervisor (QEMU).
5. Avoid touching the VirtIO network device that is providing access to your session.
6. Never attempt to read or write 0x0 as you will trigger an exception.
7. In fact never attempt to read addresses which are known to trigger an exception.

Think step by step and use the tools iteratively.`

	user := "Begin autonomous security analysis of QEMU microvm I/O space. Explore freely."

	ctx := context.Background()
	fmt.Println("[kanzashi] starting agentic microvm audit...")

	if err := runAgent(ctx, &client, system, user); err != nil {
		log.Fatal(err)
	}
}
