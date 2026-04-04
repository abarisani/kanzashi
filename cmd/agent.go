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

	shell.Add(shell.Cmd{
		Name:    "prompt",
		Args:    2,
		Pattern: regexp.MustCompile(`^prompt (system|user)(?: (.*))?$`),
		Help:    "show/change prompt",
		Syntax:  "(system|user) (<text>)?",
		Fn:      promptCmd,
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

func promptCmd(_ *shell.Interface, arg []string) (res string, err error) {
	switch arg[0] {
	case "system":
		if len(arg[1]) == 0 {
			return system, nil
		} else {
			system = arg[1]
		}
	case "user":
		if len(arg[1]) == 0 {
			return user, nil
		} else {
			user = arg[1]
		}
	}

	return
}
