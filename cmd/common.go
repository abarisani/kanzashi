// Copyright (c) The kanzashi authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/pprof"

	"github.com/usbarmory/go-boot/shell"
)

var (
	Banner   string
	Resolver = "8.8.8.8:53"
)

func init() {
	Banner = fmt.Sprintf("kanzashi • %s/%s (%s)",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	shell.Add(shell.Cmd{
		Name: "build",
		Help: "build information",
		Fn:   buildInfoCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "exit,quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "exit application",
		Fn:      exitCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stack",
		Help: "goroutine stack trace (current)",
		Fn:   stackCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stackall",
		Help: "goroutine stack trace (all)",
		Fn:   stackallCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "dns",
		Args:    1,
		Pattern: regexp.MustCompile(`^dns (.*)`),
		Syntax:  "<host>",
		Help:    "resolve domain",
		Fn:      dnsCmd,
	})

	net.SetDefaultNS([]string{Resolver})
}

func buildInfoCmd(_ *shell.Interface, _ []string) (string, error) {
	res := new(bytes.Buffer)

	if bi, ok := debug.ReadBuildInfo(); ok {
		res.WriteString(bi.String())
	}

	return res.String(), nil
}

func exitCmd(_ *shell.Interface, _ []string) (res string, err error) {
	return "", io.EOF
}

func stackCmd(_ *shell.Interface, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *shell.Interface, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func dnsCmd(_ *shell.Interface, arg []string) (res string, err error) {
	cname, err := net.LookupHost(arg[0])

	if err != nil {
		return "", fmt.Errorf("query error: %v", err)
	}

	return fmt.Sprintf("%+v\n", cname), nil
}
