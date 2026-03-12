//go:build microvm

package network

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/ioapic"
	"github.com/usbarmory/virtio-net"

	_ "golang.org/x/crypto/x509roots/fallback"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

var (
	MAC      = "1a:55:89:a2:69:41"
	Netmask  = "255.255.255.0"
	IP       = "10.0.0.1"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
)

func init() {
	microvm.AMD64.SetTime(kvmclock.Now().UnixNano())
}

func Start() (err error) {
	iface := vnet.Interface{}

	dev := &vnet.Net{
		Transport: &virtio.MMIO{
			Base: microvm.VIRTIO_NET0_BASE,
		},
		IRQ:          microvm.VIRTIO_NET0_IRQ,
		HeaderLength: 10,
	}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Socket

	dev.Start(false)
	startInterruptHandler(dev, microvm.AMD64, microvm.IOAPIC1)

	return
}

func startInterruptHandler(dev *vnet.Net, cpu *amd64.CPU, ioapic *ioapic.IOAPIC) {
	if dev == nil {
		log.Fatal("invalid device")
	}

	if cpu.LAPIC != nil {
		cpu.LAPIC.Enable()
	}

	if ioapic != nil {
		ioapic.EnableInterrupt(dev.IRQ, vector)
	}

	isr := func(irq int) {
		switch irq {
		case vector:
			for buf := dev.Rx(); buf != nil; buf = dev.Rx() {
				dev.RxHandler(buf)
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
