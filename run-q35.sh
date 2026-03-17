set -x
# download and compile GOOS=tamago compiler
TAMAGO=$(go tool -n github.com/usbarmory/tamago/cmd/tamago)
GOOS=tamago GOARCH=amd64 GOOSPKG=github.com/usbarmory/tamago $TAMAGO build \
  -tags $LLM,q35,semihosting -trimpath \
  -ldflags "-T 0x10010000 -R 0x1000 -X 'github.com/usbarmory/kanzashi/internal/llm.ClaudeAPIKey=${CLAUDE_API_KEY}' -X 'github.com/usbarmory/kanzashi/internal/llm.GeminiAPIKey=${GEMINI_API_KEY}'" *.go && \
qemu-system-x86_64 \
  -machine q35,pit=off,pic=off \
  -smp 1 \
  -global virtio-mmio.force-legacy=false \
  -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
  -m 4G -nographic -monitor none -serial stdio \
  -device pcie-root-port,port=0x10,chassis=1,id=pci.0,bus=pcie.0,multifunction=on,addr=0x3 \
  -device virtio-net-pci,netdev=net0,mac=42:01:0a:84:00:02,disable-modern=true -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
  -device virtio-gpu-pci \
  -kernel main
