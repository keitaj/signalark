# signalark

GNSS signal data collection pipeline. Connects to a u-blox receiver, captures raw UBX binary, and outputs structured CSVs for downstream analysis.

Built on [go-ubx](https://github.com/keitaj/go-ubx).

## Features

- Raw UBX binary recording with auto-generated timestamped filenames
- File rotation (`-rotate 1h` for hourly splits)
- Structured CSV output for NAV-PVT, NAV-SIG, MON-RF, RXM-RAWX
- Metadata recording (port, baud rate, location, antenna)
- Serial port auto-detection (Linux `/dev/ttyACM*`, macOS `/dev/cu.usbmodem*`)
- Receiver auto-configuration (enables NAV-PVT, NAV-SIG, RXM-RAWX, MON-RF, RXM-SFRBX)

## Usage

```bash
# Structured collection with CSV output
signalark -port /dev/ttyACM0 -dir ./collect -csv

# With file rotation and metadata
signalark -dir ./collect -csv -rotate 1h -location "Seya, Yokohama" -antenna "ANN-MB-00"

# Simple raw capture (ubxcap-compatible)
signalark -port /dev/ttyACM0 -out capture.ubx

# Auto-detect port, 10Hz, quiet mode
signalark -dir ./collect -csv -rate 100 -quiet
```

## Output Structure

When using `-dir`, signalark creates:

```
collect/
├── raw/
│   ├── gnss_20260407_120000.ubx   # Raw binary (rotated if -rotate set)
│   └── gnss_20260407_130000.ubx
├── parsed/
│   ├── nav_pvt.csv                # Position, velocity, time
│   ├── nav_sig.csv                # Per-signal quality and health
│   ├── mon_rf.csv                 # RF/jamming indicators
│   └── rxm_rawx.csv              # Raw measurements
└── metadata.json                  # Collection conditions
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | auto-detect | Serial port path |
| `-baud` | 115200 | Baud rate |
| `-rate` | 1000 | Measurement interval (ms) |
| `-dir` | | Output directory (enables structured output) |
| `-out` | | Raw file path (mutually exclusive with `-dir`) |
| `-csv` | false | Enable CSV output (requires `-dir`) |
| `-rotate` | | File rotation interval (e.g., `1h`, `30m`) |
| `-quiet` | false | Suppress console output |
| `-location` | | Collection location (metadata) |
| `-antenna` | | Antenna description (metadata) |
| `-notes` | | Additional notes (metadata) |

## Install

```
go install github.com/keitaj/signalark@latest
```

## License

MIT
