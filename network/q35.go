//go:build q35

package network

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/google/gcp"
	"github.com/usbarmory/tamago/kvm/clock"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/ioapic"
	"github.com/usbarmory/tamago/soc/intel/pci"
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
	gcp.AMD64.SetTime(kvmclock.Now().UnixNano())
}

func Start() (err error) {
	iface := vnet.Interface{}

	transport := &virtio.LegacyPCI{
		Device: pci.Probe(
			0,
			gcp.VIRTIO_NET_PCI_VENDOR,
			gcp.VIRTIO_NET_PCI_DEVICE,
		),
	}

	dev := &vnet.Net{
		Transport:    transport,
		IRQ:          vector,
		HeaderLength: 10,
	}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Socket

	gcp.AMD64.ClearInterrupt()
	dev.Start(false)
	transport.EnableInterrupt(vector, vnet.ReceiveQueue)
	startInterruptHandler(dev, gcp.AMD64, gcp.IOAPIC0)

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
