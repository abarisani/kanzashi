// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8

package platform

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/board/nxp/imx8mpevk"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/go-net"

	_ "golang.org/x/crypto/x509roots/fallback"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

var (
	MAC      = "1a:55:89:a2:69:41"
	IP       = "10.0.0.1/24"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
	Terminal = imx8mpevk.UART1
	CPU      = imx8mp.ARM64
)

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint = 0x20000000 // 512MB

func StartNetwork() (err error) {
	dev := imx8mp.ENET1
	dev.MAC, _ = net.ParseMAC(MAC)

	if err := dev.Init(); err != nil {
		return fmt.Errorf("could not initialize network device, %v", err)
	}

	iface := &gnet.Interface{
		NetworkDevice: dev,
	}

	if err := iface.Init(IP, MAC, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.Stack.EnableICMP()

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Stack.Socket

	dev.Start()
	dev.EnableInterrupt(enet.IRQ_RXF)

	startInterruptHandler(dev, iface)

	return
}

func handleEthernetInterrupt(eth *enet.ENET, iface *gnet.Interface, buf []byte) {
	for {
		if n, err := eth.Receive(buf); err != nil || n == 0 {
			return
		}

		iface.Stack.RecvInboundPacket(buf)
		eth.ClearInterrupt(enet.IRQ_RXF)
	}
}

func startInterruptHandler(eth *enet.ENET, iface *gnet.Interface) {
	var buf []byte

	imx8mp.GIC.Init()
	imx8mp.GIC.EnableInterrupt(arm64.TIMER_IRQ)

	if eth != nil {
		buf = make([]byte, gnet.EthernetMaximumSize+gnet.MTU)
		imx8mp.GIC.EnableInterrupt(eth.IRQ)
	}

	isr := func() {
		irq := imx8mp.GIC.GetInterrupt()

		switch {
		case irq == arm64.TIMER_IRQ:
			imx8mp.ARM64.SetAlarm(0)
		case eth != nil && irq == eth.IRQ:
			handleEthernetInterrupt(eth, iface, buf)
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	// optimize CPU idle management as IRQs are enabled
	goos.Idle = func(pollUntil int64) {
		if pollUntil == 0 {
			return
		}

		imx8mp.ARM64.SetAlarm(pollUntil)
		imx8mp.ARM64.WaitInterrupt()
	}

	go arm64.ServiceInterrupts(isr)
}
