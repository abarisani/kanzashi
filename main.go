package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/amd64"

	"github.com/usbarmory/kanzashi/llm"
	"github.com/usbarmory/kanzashi/network"
	"github.com/usbarmory/kanzashi/tool"
)

//go:embed log.txt
var pastSession string

func init() {
	amd64.SystemExceptionHandler = nil
}

func main() {
	log.SetFlags(0)
	log.Printf("\n[kanzashi] starting network")

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
2. Focus on the IOAPIC and try to crash the VM with with malicious valyes in IRQ IRQ redirection entries.
3. Do not document anomalous hypervisor responses, just aim for a privilege escalation on the hypervisor (QEMU).
4. Avoid touching the VirtIO network device that is providing access to your session.
5. If you crash the hypervisor it is a result, if you are still alive it means you had no effect on the host.

Think step by step and use the tools iteratively.`

	user := fmt.Sprintf("Begin autonomous security analysis of QEMU VM. Explore freely. Briefly explain what you are doing as you go.")

	if len(pastSession) > 0 {
		log.Printf("[kanzashi] using past session log (%d bytes)", len(pastSession))
		user += fmt.Sprintf("Resume from the last session, here are the logs of it:%s", pastSession)
	}

	log.Printf("[kanzashi] starting agentic QEMU audit...")
	log.Printf("\n%s\n%s\n\n", system, user)

	llm.RunAgent(context.Background(), system, user)

	log.Printf("[kanzashi] graceful exit")
}
