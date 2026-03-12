set -x
# download and compile GOOS=tamago compiler
TAMAGO=$(go tool -n github.com/usbarmory/tamago/cmd/tamago)
GOOS=tamago GOARCH=amd64 GOOSPKG=github.com/usbarmory/tamago $TAMAGO build -tags $LLM,microvm,semihosting -trimpath \
  -ldflags "-T 0x10010000 -R 0x1000 -X 'github.com/usbarmory/kanzashi/llm.ClaudeAPIKey=${CLAUDE_API_KEY}' -X 'github.com/usbarmory/kanzashi/llm.GeminiAPIKey=${GEMINI_API_KEY}'" *.go && \
qemu-system-x86_64 \
  -machine microvm,x-option-roms=on,pit=off,pic=off,rtc=on \
  -smp 1 \
  -global virtio-mmio.force-legacy=false \
  -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
  -m 4G -nographic -monitor none -serial stdio \
  -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
  -kernel main
