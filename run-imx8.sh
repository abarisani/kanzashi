set -x
# download and compile GOOS=tamago compiler
TAMAGO=$(go tool -n github.com/usbarmory/tamago/cmd/tamago)
GOOS=tamago GOARCH=arm64 GOOSPKG=github.com/usbarmory/tamago $TAMAGO build \
  -tags imx8,linkramsize,semihosting -trimpath \
  -ldflags "-T 0x40010000 -R 0x1000 -X 'github.com/usbarmory/kanzashi/internal/claude.APIKey=${CLAUDE_API_KEY}' -X 'github.com/usbarmory/kanzashi/internal/gemini.APIKey=${GEMINI_API_KEY}'" *.go && \
qemu-system-aarch64 \
  -machine imx8mp-evk \
  -m 512M -smp 1 \
  -nographic -monitor none -semihosting -serial stdio -serial null \
  -net nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
  -kernel main -d guest_errors,unimp,invalid_mem,pcall
