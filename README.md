# ehw - Hardware Information Tool

A CPU information tool written in Go that displays detailed processor information in a retro-style TUI interface.

## Screenshots

The application presents hardware information in a clean, retro-styled terminal interface with box-drawing characters and color-coded sections.

## Features

- **Summary Page**: Overview of CPU information including vendor, brand, cores, threads, features count, and cache summary
- **CPU Page**: Comprehensive CPU details including:
  - Basic information (vendor, brand, model, family, stepping)
  - Core and thread counts
  - CPUID function information
  - Physical and linear address bits
  - Processor details (logical processors, APIC ID, threads per core)
  - Model data (stepping, model, family IDs)
  - Hybrid CPU detection (Intel P-core/E-core)
  - Detailed cache information (L1, L2, L3 with associativity, line size, sets)
  - TLB (Translation Lookaside Buffer) information
  - Supported CPU features organized by category

## Navigation

| Key | Action |
|-----|--------|
| `←` `→` | Navigate between pages |
| `↑` `↓` | Scroll content |
| Mouse Wheel | Scroll content |
| Mouse Click | Select menu items |
| `Q` | Quit the application |
| `Ctrl+C` / `Esc` | Quit the application |

## Requirements

- Go 1.24 or later
- Terminal with TUI support
- x86/x64 or ARM64 processor

## Installation

```bash
go build -o ehw
```

Or install directly:

```bash
go install github.com/earentir/ehw@latest
```

## Usage

```bash
./ehw
```

## Dependencies

- [tcell](https://github.com/gdamore/tcell) - Terminal cell library for TUI
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [retrotui](https://github.com/earentir/retrotui) - Retro-style TUI library
- [cpuid](https://github.com/earentir/cpuid) - Comprehensive CPU identification library

## Platform Support

| Platform | Support |
|----------|---------|
| Linux (x86/x64) | ✅ Full support |
| macOS (x86/x64) | ✅ Full support |
| macOS (ARM64/Apple Silicon) | ✅ Full support |
| Windows (x86/x64) | ✅ Full support |
| Linux (ARM64) | ✅ Full support |

### Supported Information by Platform

| Information | x86/x64 | ARM64/Apple Silicon |
|-------------|:-------:|:-------------------:|
| **Basic Information** |
| Vendor Name | ✅ | ✅ |
| Brand String | ✅ | ✅ |
| **Model Data** |
| Family ID | ✅ | ✅ |
| Model ID | ✅ | ✅ |
| Stepping ID | ✅ | ✅ |
| Extended Family | ✅ | ✅ |
| Extended Model | ✅ | ✅ |
| Processor Type | ✅ | ✅ |
| **Processor Details** |
| Core Count | ✅ | ✅ |
| Thread Count | ✅ | ✅ |
| Threads per Core | ✅ | ✅ |
| Max Logical Processors | ✅ | ✅ |
| Physical Address Bits | ✅ | ✅ |
| Linear Address Bits | ✅ | ✅ |
| **Cache Information** |
| L1 Data Cache | ✅ | ✅ |
| L1 Instruction Cache | ✅ | ✅ |
| L2 Cache | ✅ | ✅ |
| L3 Cache | ✅ | ⚠️ (if present) |
| Cache Associativity | ✅ | ✅ |
| Cache Line Size | ✅ | ✅ |
| **TLB Information** |
| TLB Entries | ✅ | ❌ |
| **Hybrid/Heterogeneous CPU** |
| P-core/E-core Detection | ✅ (Intel) | ✅ (Apple Silicon) |
| **CPU Features** |
| Feature Detection | ✅ | ✅ |
| Feature Categories | ✅ | ✅ |
| x86 Features (SSE, AVX, etc.) | ✅ | N/A |
| ARM Features (NEON, ASIMD, etc.) | N/A | ✅ |

### ARM64/Apple Silicon Features

On Apple Silicon (M1/M2/M3), the following feature categories are detected:

- **SIMD**: NEON, ASIMD, ASIMD_HP, ASIMD_DP, ASIMDFHM
- **Cryptography**: AES, SHA1, SHA256, PMULL
- **Floating Point**: FP, FP16, FHM, FRINTTS
- **Memory**: ATOMICS, DPB, DPB2, LRCPC
- **Security**: SSBS, BTI, DIT, SB
- **Other**: CRC32, FCMA, JSCVT, AMX

## License

GNU General Public License v2.0 - see [LICENSE](LICENSE) file for details.
