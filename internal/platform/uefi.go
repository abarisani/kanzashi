// Copyright (c) The kanzashi Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build uefi

package platform

import (
	"fmt"
	"net"

	"github.com/usbarmory/tamago/soc/intel/pci"
	"github.com/usbarmory/tamago/kvm/virtio"

	"github.com/usbarmory/go-boot/uefi/x64"

	"github.com/usbarmory/virtio-net"

	_ "golang.org/x/crypto/x509roots/fallback"
)

const (
	VIRTIO_NET_PCI_VENDOR = 0x1af4 // Red Hat, Inc.

	// Virtio 1.0 network device
	VIRTIO_NET_PCI_LEGACY_DEVICE = 0x1000
	VIRTIO_NET_PCI_MODERN_DEVICE = 0x1041
)

var (
	MAC      = "1a:55:89:a2:69:41"
	Netmask  = "255.255.255.0"
	IP       = "10.0.0.1"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
	Terminal = x64.UART0
)

func init() {
	x64.UEFI.Boot.SetWatchdogTimer(0)
	x64.AllocateDMA(2 << 20)
}

func StartNetwork() (err error) {
	var nic *vnet.Net
	iface := vnet.Interface{}

	if device := pci.Probe(
		0,
		VIRTIO_NET_PCI_VENDOR,
		VIRTIO_NET_PCI_LEGACY_DEVICE,
	); device != nil {
		nic = &vnet.Net{
			Transport: &virtio.LegacyPCI{
				Device: device,
			},
			HeaderLength: 10,
		}
	} else if device := pci.Probe(
		0,
		VIRTIO_NET_PCI_VENDOR,
		VIRTIO_NET_PCI_MODERN_DEVICE,
	); device != nil {
		nic = &vnet.Net{
			Transport: &virtio.PCI{
				Device: device,
			},
			HeaderLength: 10,
		}
	}

	if err := iface.Init(nic, IP, Netmask, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()
	go nic.Start(true)

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Socket

	nic.Start(false)

	return
}
