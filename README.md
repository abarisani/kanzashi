Introduction
============

This repository experiments with bare metal Go, as supported by
[TamaGo](https://github.com/usbarmory/tamago), to perform LLM-driven security
vulnerability analysis on low-level architectural layers minimizing OS
interference.

The TamaGo framework allows trivial integration of LLM libraries on bare metal
to achieve the following:

  * dramatically reduced driver footprint, no OS presence
  * ability to import and run pure Go libraries
  * ability to bridge register/MSR read/write tooling with an LLM agent

This example targets a QEMU `microvm` machine, it can be easily expanded to any
`amd64`, `arm`, `arm64`, `riscv64` target as supported by
[TamaGo](https://github.com/usbarmory/tamago).

This effort is meant to show the effectiveness of TamaGo in bridging layers
which are often considered far apart, rather than making any statements or
promises on the effectiveness of wan LLM-drive approach in testing QEMU.

PRs are welcome to add success stories/bugs found with this approach.

Usage
=====

1. Download and patch the [TamaGo](https://github.com/usbarmory/tamago) library
   with the `tamago.patch` in this repository, the patch allows to continue
   execution on register read/write faults and skip the faulty instruction.

2. Point `go.mod` to the patched library with either a
   [Go workspace](https://go.dev/blog/get-familiar-with-workspaces)
   or `go mod edit -replace`.

3. Adapt the system and user prompts in `main.go`, copy a past `log.txt` or
   `touch` an empty one to begin from scratch.

4. Change, if desired, the default model selected in `internal/llm.ClaudeModel`
   or `internal/llm.GeminiModel`.

4. Ensure `CLAUDE_API_KEY` and/or `GEMINI_API_KEY` are set in your environment.

5. Launch either `run-microvm.sh` or `run-q35.sh` for an example QEMU
   configuration.

Example (Claude)
================

The following example shows discovery of [QEMU #3336](https://gitlab.com/qemu-project/qemu/-/work_items/3336):

```
./run-microvm.sh

...

[kanzashi] checking exception handler patch
[kanzashi] read32 0x0 => 0x0 (fault)
[kanzashi] starting network

kanzashi • tamago/amd64

agent           (claude|gemini)? # start agent
build                            # build information
dns             <host>           # resolve domain
exit,quit                        # exit application
help                             # this help
prompt   (system|user) (<text>)? # show/change prompt
stack                            # goroutine stack trace (current)
stackall                         # goroutine stack trace (all)

[kanzashi] starting agentic QEMU audit...
[kanzashi] initializing claude agent (claude-opus-4-6)

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

Returns 0 for out-of-range registers. Let me try writing to
redirection entries with malicious values. The redirection table starts at
register 0x10:

[kanzashi] write32 0xfec00000 <= 0x10 (<nil>)
[kanzashi] write32 0xfec00010 <= 0xffffffff

Segmentation fault (core dumped)

Interesting! The previous log shows that writing `0xFFFFFFFF` to
IOAPIC redirection table entry 0x10 via the IOWIN register caused a segfault in
QEMU. That's already a successful crash.
```

Author
======

Andrea Barisani  
andrea@inversepath.com  

License
=======

Copyright (c) The kanzashi authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/kotama/blob/main/LICENSE) file.
