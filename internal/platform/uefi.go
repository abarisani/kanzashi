// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build uefi

package platform

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"

	"github.com/usbarmory/tamago/soc/intel/pci"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/ioapic"

	"github.com/usbarmory/go-boot/uefi/x64"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"

	_ "golang.org/x/crypto/x509roots/fallback"
)

const (
	IOAPIC0_BASE = 0xfec00000

	VIRTIO_NET_PCI_VENDOR = 0x1af4 // Red Hat, Inc.

	// Virtio 1.0 network device
	VIRTIO_NET_PCI_LEGACY_DEVICE = 0x1000
	VIRTIO_NET_PCI_MODERN_DEVICE = 0x1041

	// redirection vector for IOAPIC IRQ to CPU IRQ or MSI-X signal
	VIRTIO_NET_IRQ = 32
)

var (
	MAC      = "1a:55:89:a2:69:41"
	IP       = "10.0.0.1/24"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
	Terminal = x64.UART0
)

func init() {
	x64.UEFI.Boot.SetWatchdogTimer(0)
	x64.AllocateDMA(2 << 20)
}

func StartNetwork() (err error) {
	nic := &vnet.Net{
		IRQ:          VIRTIO_NET_IRQ,
		MTU:          gnet.MTU,
		HeaderLength: 10,
	}

	if dev := pci.Probe(
		0,
		VIRTIO_NET_PCI_VENDOR,
		VIRTIO_NET_PCI_LEGACY_DEVICE,
	); dev != nil {
		nic.Transport = &virtio.LegacyPCI{
			Device: dev,
		}
	} else if dev := pci.Probe(
		0,
		VIRTIO_NET_PCI_VENDOR,
		VIRTIO_NET_PCI_MODERN_DEVICE,
	); dev != nil {
		nic.Transport = &virtio.PCI{
			Device: dev,
		}
	}

	if err := nic.Init(); err != nil {
		return fmt.Errorf("could not initialize VirtIO device, %v", err)
	}

	iface := &gnet.Interface{
		NetworkDevice: nic,
	}

	if err := iface.Init(IP, MAC, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.Stack.EnableICMP()

	go func() {
		nic.Transport.EnableInterrupt(nic.IRQ, vnet.ReceiveQueue)
		startInterruptHandler(nic, iface)
	}()

	nic.Start()

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Stack.Socket

	return
}

func startInterruptHandler(dev *vnet.Net, iface *gnet.Interface) {
	if dev == nil || iface == nil {
		log.Fatal("invalid device")
	}

	cpu := x64.AMD64
	cpu.EnableExceptions()

	if cpu.LAPIC != nil {
		cpu.LAPIC.Enable()
	}

	ioapic := &ioapic.IOAPIC{
		Base: IOAPIC0_BASE,
	}

	ioapic.EnableInterrupt(dev.IRQ, dev.IRQ)

	size := dev.HeaderLength + gnet.EthernetMaximumSize + gnet.MTU
	buf := make([]byte, size)

	isr := func(irq int) {
		switch irq {
		case dev.IRQ:
			for {
				if n, err := dev.ReceiveWithHeader(buf); err != nil || n == 0 {
					return
				}

				iface.Stack.RecvInboundPacket(buf[dev.HeaderLength:])
			}
		default:
			log.Printf("internal error, unexpected IRQ %d (%d)", irq)
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

	cpu.ServiceInterrupts(isr)
}
