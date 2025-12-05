# earhw - Hardware Information Tool

A hardware information tool written in Go that displays CPU, RAM, and disk information in a retro-style TUI interface.

## Features

- **Summary Page**: Overview of all hardware components
- **CPU Page**: Detailed CPU information including vendor, brand, cores, threads, cache, and supported features
- **RAM Page**: Memory information with usage visualization
- **Disk Page**: Storage device information including partitions

## Navigation

- **Left/Right Arrow Keys**: Navigate between pages (Summary, CPU, RAM, Disk)
- **Q**: Quit the application
- **Ctrl+C**: Quit the application

## Requirements

- Go 1.24 or later
- Terminal with TUI support

## Installation

```bash
go build -o earhw
```

## Usage

```bash
./earhw
```

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [retrotui](https://github.com/earentir/retrotui) - Retro-style TUI library
- [cpuid](https://github.com/earentir/cpuid) - CPU information (x86/x64 platforms)
- [ghw](https://github.com/jaypipes/ghw) - Hardware information library

## Platform Support

- **macOS**: Memory information via sysctl, CPU and disk via ghw
- **Linux**: Memory, CPU, and disk information via ghw (cpuid attempted for x86/x64)
- **Windows**: Memory, CPU, and disk information via ghw (cpuid attempted for x86/x64)

The application automatically detects the platform and uses the appropriate methods for hardware information collection.

## License

This project uses the same license as its dependencies.
