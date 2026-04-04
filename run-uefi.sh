set -x
# download and compile GOOS=tamago compiler
TAMAGO=$(go tool -n github.com/usbarmory/tamago/cmd/tamago)
GOOS=tamago GOARCH=amd64 GOOSPKG=github.com/usbarmory/tamago $TAMAGO build \
  -tags uefi,linkcpuinit,linkramstart,linkramsize,linkprintk,semihosting -trimpath \
  -ldflags "-E cpuinit -T 0x10010000 -R 0x1000 -X 'github.com/usbarmory/kanzashi/internal/claude.APIKey=${CLAUDE_API_KEY}' -X 'github.com/usbarmory/kanzashi/internal/gemini.APIKey=${GEMINI_API_KEY}'" *.go && \

objcopy \
  --strip-debug \
  --output-target efi-app-x86_64 \
  --subsystem=efi-app \
  --image-base 0x10000000 \
  --stack=0x10000 \
  main main.efi && \
  printf '\x26\x02' | dd of=main.efi bs=1 seek=150 count=2 conv=notrunc,fsync && \

mkdir -p $PWD/qemu-disk/efi/boot && cp $PWD/main.efi $PWD/qemu-disk/efi/boot/bootx64.efi && \

OVMFCODE="OVMF_CODE.fd"
qemu-system-x86_64 -machine q35,pit=off,pic=off \
        -m 4G -smp 1 \
        -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
        -device pcie-root-port,port=0x10,chassis=1,id=pci.0,bus=pcie.0,multifunction=on,addr=0x3 \
        -device virtio-net-pci,netdev=net0,mac=42:01:0a:84:00:02,disable-modern=true -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
        -drive format=raw,file=fat:rw:$PWD/qemu-disk \
        -drive if=pflash,format=raw,readonly,file=$OVMFCODE \
        -global isa-debugcon.iobase=0x402 \
        -serial stdio -nographic -monitor none  \
