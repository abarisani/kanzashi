
Example (Claude)
================

```
[kanzashi] instantiating claude client
[kanzashi] starting agentic microvm audit...

[claude]

I'll begin by systematically exploring the QEMU microvm I/O space. The microvm
machine type has a minimal set of peripherals. Let me start by probing known
address ranges for the microvm platform.

## Step 1: Probe known microvm MMIO regions

The microvm uses a simplified memory map. Key regions include:
- **LAPIC**: 0xFEE00000
- **IOAPIC**: 0xFEC00000
- **HPET**: 0xFED00000
- **VirtIO MMIO**: typically at 0xFEB00000+ (each device gets a 0x200 region)
- **fw_cfg**: 0xFEB00000 range or I/O port mapped

[claude] Returns 0 for out-of-range registers. Let me try writing to
redirection entries with malicious values. The redirection table starts at
register 0x10:

write fec00000 < 10
write fec00010 < ffffffff

Segmentation fault (core dumped)

[claude] Interesting! The previous log shows that writing `0xFFFFFFFF` to
IOAPIC redirection table entry 0x10 via the IOWIN register caused a segfault in
QEMU. That's already a successful crash.
```
