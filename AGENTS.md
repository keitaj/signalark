This file provides guidance to AI agents when working with code in the signalark repository.

## Project Overview

signalark is a GNSS signal data collection pipeline written in Go. It connects to a u-blox receiver via serial port, captures raw UBX binary data, and outputs structured CSVs for downstream analysis and ML training.

Built on [go-ubx](https://github.com/keitaj/go-ubx) for UBX protocol parsing.

## Principles

- Follow existing patterns in surrounding code
- Keep changes focused — avoid over-engineering
- Security: validate user input at system boundaries
- Platform-aware: serial port handling differs between macOS and Linux

## Commands

### Build

```bash
go build ./...          # Build all packages
go build -o signalark   # Build binary
go vet ./...            # Static analysis
```

### Install

```bash
go install github.com/keitaj/signalark@latest
```

### Run

```bash
# Requires a u-blox GNSS receiver connected via USB
signalark -dir ./collect/session_001 -csv \
  -mobility static -skyvis open \
  -notes "test session"
```

## Architecture

Single-package (`main`) CLI application. All source files are in the project root.

| File               | Purpose                                          |
|--------------------|--------------------------------------------------|
| `main.go`          | Entry point, flag parsing, orchestration         |
| `config.go`        | UBX receiver configuration (CFG-VALSET + ACK)    |
| `recorder.go`      | Raw binary file writing with optional rotation   |
| `csv.go`           | CSV writers for all supported UBX message types  |
| `metadata.go`      | Session metadata (conditions, labels, position)  |
| `messages.go`      | Message type selection and validation            |
| `display.go`       | Console output formatting                        |
| `status.go`        | Running statistics (epochs, fix rate, CN0)       |
| `gaps.go`          | Data gap detection and logging                   |
| `serial_darwin.go` | macOS serial port baud rate mapping              |
| `serial_linux.go`  | Linux serial port baud rate mapping              |
| `serial_unix.go`   | Unix serial port open/configure (8N1 raw mode)   |

### Data Flow

1. Serial port opened and configured (platform-specific)
2. CFG-VALSET sent to receiver to enable selected UBX messages
3. Decode loop: UBX binary → TeeReader branches to raw recorder + CSV writers
4. Periodic flush (10s) syncs all files and prints status
5. Graceful shutdown on SIGINT/SIGTERM

### Supported UBX Messages

| Message      | CSV File         | Content                          |
|--------------|------------------|----------------------------------|
| NAV-PVT      | `nav_pvt.csv`    | Position, velocity, time         |
| NAV-SAT      | `nav_sat.csv`    | Per-satellite elevation, azimuth |
| NAV-SIG      | `nav_sig.csv`    | Per-signal quality and health    |
| MON-RF       | `mon_rf.csv`     | RF/jamming indicators            |
| RXM-RAWX     | `rxm_rawx.csv`   | Raw measurements                 |
| RXM-SFRBX    | `rxm_sfrbx.csv`  | Navigation subframe data         |

## Key Patterns

- **Adding a new UBX message type**: Update `messages.go` (MessageSet, validMessages, parseMessages, Names), `config.go` (CFG-VALSET key), `csv.go` (new CSV writer struct + integrate into csvWriters), and `main.go` (default flag value).
- **CSV writers**: Each writer is a struct with `*csv.Writer` and `*os.File`. Use `bufio.NewWriter(f)` when creating the `csv.Writer`. Follow the existing `newXxxCSV` / `Write` / `Close` pattern.
- **Error cleanup in `newCSVWriters`**: On error, close all previously opened writers before returning.
- **Platform serial code**: Build-tagged files (`serial_darwin.go`, `serial_linux.go`) provide baud rate constants. Shared logic lives in `serial_unix.go`.

## Dependencies

- `github.com/keitaj/go-ubx` — UBX protocol decoder/encoder (sole external dependency)
- Go standard library only otherwise

## Output Structure

```
session/
├── raw/
│   └── gnss_YYYYMMDD_HHMMSS.ubx   # Raw binary (rotated if -rotate set)
├── parsed/
│   ├── nav_pvt.csv
│   ├── nav_sat.csv
│   ├── nav_sig.csv
│   ├── mon_rf.csv
│   ├── rxm_rawx.csv
│   ├── rxm_sfrbx.csv
│   └── gaps.csv                    # Data gap log (if gaps detected)
└── metadata.json                   # Collection conditions + session labels
```
