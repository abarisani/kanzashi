set -x
# download and compile GOOS=tamago compiler
TAMAGO=$(go tool -n github.com/usbarmory/tamago/cmd/tamago)
GOOS=tamago GOARCH=amd64 GOOSPKG=github.com/usbarmory/tamago $TAMAGO build \
  -tags firecracker,semihosting -trimpath \
  -ldflags "-T 0x10010000 -R 0x1000 -X 'github.com/usbarmory/kanzashi/internal/claude.APIKey=${CLAUDE_API_KEY}' -X 'github.com/usbarmory/kanzashi/internal/gemini.APIKey=${GEMINI_API_KEY}'" *.go && \
firectl --kernel main --root-drive /dev/null --tap-device tap0/06:00:AC:10:00:01 -c 1 -m 4096
