// Copyright (c) The kanzashi authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/usbarmory/go-boot/shell"

	"github.com/usbarmory/kanzashi/internal/claude"
	"github.com/usbarmory/kanzashi/internal/gemini"
)

const (
	system = `You are a hypervisor security researcher running bare metal inside a QEMU VM on AMD64.
You have direct access to physical memory via reg_read32, reg_write32, reg_read64, reg_write64, msr_read, msr_write tools.
Your goal is to autonomously explore the I/O peripheral range to find hypervisor implementation flaws that allow to escape the hypervisor.

Approach:
1. Assume a QEMU VM (either q35 or microvm).
2. Target the VirtIO GPU device.
3. Do not document anomalous hypervisor responses, just aim for a privilege escalation on the hypervisor (QEMU).
4. Avoid touching the VirtIO network device that is providing access to your session.
5. Your messages are sent on a VT100 compatible UART, so use colors to categorize or highlight output accordingly.

Think step by step and use the tools iteratively.`

	user = `Begin autonomous security analysis of QEMU VM. Explore freely. Briefly explain what you are doing as you go.`
)

var PastSession string

func init() {
	shell.Add(shell.Cmd{
		Name:    "agent",
		Args:    1,
		Pattern: regexp.MustCompile(`^agent(?: (claude|gemini))?$`),
		Help:    "start agent",
		Syntax:  "(claude|gemini)?",
		Fn:      agentCmd,
	})
}

func agentCmd(_ *shell.Interface, arg []string) (res string, err error) {
	userPrompt := user

	if len(PastSession) > 0 {
		log.Printf("[kanzashi] using past session log (%d bytes)", len(PastSession))
		userPrompt += fmt.Sprintf("Here are the logs of the past sessions, resume from them:%s", PastSession)
	}

	log.Printf("[kanzashi] starting agentic QEMU audit...")
	log.Printf("\n%s\n%s\n\n", system, userPrompt)

	switch arg[0] {
	case "claude":
		go claude.RunAgent(context.Background(), system, userPrompt)
	case "gemini":
		go gemini.RunAgent(context.Background(), system, userPrompt)
	default:
		go gemini.RunAgent(context.Background(), system, userPrompt)
	}

	return "started", nil
}
