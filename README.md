# signalark

GNSS signal data collection pipeline. Connects to a u-blox receiver, captures raw UBX binary, and outputs structured CSVs for downstream analysis.

Built on [go-ubx](https://github.com/keitaj/go-ubx).

## Features

- Raw UBX binary recording with auto-generated timestamped filenames
- File rotation (`-rotate 1h` for hourly splits)
- Structured CSV output for NAV-PVT, NAV-SIG, MON-RF, RXM-RAWX
- Session labels for ML training data (`-mobility`, `-skyvis`, `-weather`, `-anomaly`)
- Auto-recorded start position (lat/lon from first NAV-PVT fix)
- Metadata recording (`metadata.json`)
- Serial port auto-detection (Linux `/dev/ttyACM*`, macOS `/dev/cu.usbmodem*`)
- Receiver auto-configuration (enables NAV-PVT, NAV-SIG, RXM-RAWX, MON-RF, RXM-SFRBX)

## Usage

```bash
# Static observation, open sky
signalark -dir ./collect/session_001 -csv \
  -mobility static -skyvis open \
  -antenna patch -notes "Seya Park, clear sky, tripod"

# Walking in urban area
signalark -dir ./collect/session_002 -csv \
  -mobility walk -skyvis urban \
  -antenna patch -notes "Yokohama Station west exit"

# Driving with roof antenna
signalark -dir ./collect/session_003 -csv \
  -mobility drive -skyvis open \
  -antenna magnet_roof -notes "Hodogaya Bypass"

# Simple raw capture (no session labels)
signalark -port /dev/ttyACM0 -out capture.ubx

# Auto-detect port, 10Hz, quiet mode
signalark -dir ./collect -csv -rate 100 -quiet
```

## Output Structure

When using `-dir`, signalark creates:

```
session_001/
├── raw/
│   ├── gnss_20260407_120000.ubx   # Raw binary (rotated if -rotate set)
│   └── gnss_20260407_130000.ubx
├── parsed/
│   ├── nav_pvt.csv                # Position, velocity, time
│   ├── nav_sig.csv                # Per-signal quality and health
│   ├── mon_rf.csv                 # RF/jamming indicators
│   └── rxm_rawx.csv              # Raw measurements
└── metadata.json                  # Collection conditions + session labels
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
| `-antenna` | | Antenna description |
| `-mobility` | | Mobility mode: `static`, `walk`, `drive` |
| `-skyvis` | | Sky visibility: `open`, `suburban`, `urban`, `canyon`, `indoor`, `tunnel` |
| `-weather` | | Weather: `clear`, `cloudy`, `rain`, `snow` |
| `-anomaly` | `normal` | Anomaly label: `normal`, `spoofing`, `jamming` |
| `-notes` | | Free-form notes (location name, conditions, etc.) |

## Install

```
go install github.com/keitaj/signalark@latest
```

## License

MIT
