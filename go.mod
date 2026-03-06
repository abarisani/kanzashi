module github.com/usbarmory/kanzashi

go 1.26.0

tool github.com/usbarmory/tamago/cmd/tamago

replace github.com/usbarmory/tamago => /mnt/git/public/tamago

require (
	github.com/anthropics/anthropic-sdk-go v1.26.0
	github.com/usbarmory/tamago v1.26.0
	github.com/usbarmory/virtio-net v0.0.0-20250916125519-733a429bd100
	golang.org/x/crypto/x509roots/fallback v0.0.0-20260213171211-a408498e5541
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	gvisor.dev/gvisor v0.0.0-20250115195935-26653e7d8816 // indirect
)
