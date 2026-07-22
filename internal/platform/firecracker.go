// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker

package platform

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/firecracker/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/ioapic"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"

	_ "golang.org/x/crypto/x509roots/fallback"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

var (
	MAC      = "1a:55:89:a2:69:41"
	IP       = "10.0.0.1/24"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
	Terminal = microvm.UART0
	CPU      = microvm.AMD64
)

func init() {
	microvm.AMD64.SetTime(kvmclock.Now().UnixNano())
}

func StartNetwork() (err error) {
	dev := &vnet.Net{
		Transport: &virtio.MMIO{
			Base: microvm.VIRTIO_NET0_BASE,
		},
		IRQ:          microvm.VIRTIO_NET0_IRQ,
		MTU:          gnet.MTU,
	}

	if err := dev.Init(); err != nil {
		return fmt.Errorf("could not initialize VirtIO device, %v", err)
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
	startInterruptHandler(dev, iface, microvm.AMD64, microvm.IOAPIC0)

	return
}

func startInterruptHandler(dev *vnet.Net, iface *gnet.Interface, cpu *amd64.CPU, ioapic *ioapic.IOAPIC) {
	if dev == nil {
		log.Fatal("invalid device")
	}

	if cpu.LAPIC != nil {
		cpu.LAPIC.Enable()
	}

	if ioapic != nil {
		ioapic.EnableInterrupt(dev.IRQ, vector)
	}

	size := dev.HeaderLength + gnet.EthernetMaximumSize + gnet.MTU
	buf := make([]byte, size)

	isr := func(irq int) {
		switch irq {
		case vector:
			for {
				if n, err := dev.ReceiveWithHeader(buf); err != nil || n == 0 {
					return
				}

				iface.Stack.RecvInboundPacket(buf[dev.HeaderLength:])
			}
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	// optimize CPU idle management as IRQs are enabled
	goos.Idle = func(pollUntil int64) {
		if pollUntil == 0 {
			return
		}

		cpu.SetAlarm(pollUntil)
		cpu.WaitInterrupt()
		cpu.SetAlarm(0)
	}

	go cpu.ServiceInterrupts(isr)
}
