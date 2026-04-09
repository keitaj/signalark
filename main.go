// signalark connects to a u-blox GNSS receiver and captures signal data
// to raw binary files and structured CSVs for downstream analysis.
//
// Usage:
//
//	signalark -dir ./collect -csv -mobility static -skyvis open
//	signalark -dir ./collect -csv -rotate 1h -mobility walk -skyvis urban
//	signalark -port /dev/ttyACM0 -out capture.ubx
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// nowUnixMs returns the current time as Unix milliseconds.
// Defined as a package-level function so csv.go can use it.
func nowUnixMs() int64 { return time.Now().UnixMilli() }

func main() {
	portName := flag.String("port", "", "Serial port (e.g., /dev/ttyACM0, /dev/cu.usbmodem*)")
	baudRate := flag.Int("baud", 115200, "Baud rate")
	measRate := flag.Int("rate", 1000, "Measurement interval in ms (1000=1Hz, 100=10Hz)")
	outFile := flag.String("out", "", "Output file for raw UBX binary (mutually exclusive with -dir)")
	outDir := flag.String("dir", "", "Output directory for structured data (raw/, parsed/, metadata.json)")
	csvFlag := flag.Bool("csv", false, "Enable CSV output (requires -dir)")
	rotateStr := flag.String("rotate", "", "File rotation interval (e.g., 1h, 30m)")
	quiet := flag.Bool("quiet", false, "Suppress console output")
	antenna := flag.String("antenna", "", "Antenna description (metadata)")
	mobility := flag.String("mobility", "", "Mobility mode: static, walk, drive")
	skyvis := flag.String("skyvis", "", "Sky visibility: open, suburban, urban, canyon, indoor, tunnel")
	weather := flag.String("weather", "", "Weather: clear, cloudy, rain, snow")
	anomaly := flag.String("anomaly", "normal", "Anomaly label: normal, spoofing, jamming")
	notes := flag.String("notes", "", "Free-form notes (e.g., location name, conditions)")
	itfm := flag.Bool("itfm", false, "Enable ITFM interference monitor (requires firmware support)")
	flag.Parse()

	if *measRate <= 0 {
		log.Fatalf("Invalid measurement rate: %d ms (must be > 0)", *measRate)
	}
	if *outFile != "" && *outDir != "" {
		log.Fatal("-out and -dir are mutually exclusive")
	}
	if *csvFlag && *outDir == "" {
		log.Fatal("-csv requires -dir")
	}

	var rotateDur time.Duration
	if *rotateStr != "" {
		d, err := time.ParseDuration(*rotateStr)
		if err != nil {
			log.Fatalf("Invalid -rotate value: %v", err)
		}
		rotateDur = d
	}

	// Auto-detect port
	if *portName == "" {
		*portName = detectPort()
		if *portName == "" {
			fmt.Fprintln(os.Stderr, "Usage: signalark -port /dev/ttyACM0 -dir ./collect -csv")
			fmt.Fprintln(os.Stderr, "\nTip: look for /dev/ttyACM* (Linux) or /dev/cu.usbmodem* (macOS)")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Auto-detected port: %s\n", *portName)
	}

	// Open serial port
	p, err := openPort(*portName, *baudRate)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", *portName, err)
	}
	defer p.Close()
	fmt.Fprintf(os.Stderr, "Connected to %s @ %d baud\n", *portName, *baudRate)

	// Configure receiver
	configure(p, *measRate, *itfm)

	// Set up output
	var reader io.Reader = p
	var rec *recorder
	var csvW *csvWriters
	var meta *metadata

	if *outDir != "" {
		// Create directory structure
		rawDir := filepath.Join(*outDir, "raw")
		parsedDir := filepath.Join(*outDir, "parsed")
		if err := os.MkdirAll(rawDir, 0755); err != nil {
			log.Fatalf("Failed to create %s: %v", rawDir, err)
		}
		if err := os.MkdirAll(parsedDir, 0755); err != nil {
			log.Fatalf("Failed to create %s: %v", parsedDir, err)
		}

		// Raw recorder
		rec, err = newRecorder(rawDir, rotateDur)
		if err != nil {
			log.Fatalf("Failed to create recorder: %v", err)
		}
		defer rec.Close()
		reader = io.TeeReader(p, rec)

		// CSV writers
		if *csvFlag {
			csvW, err = newCSVWriters(parsedDir)
			if err != nil {
				log.Fatalf("Failed to create CSV writers: %v", err)
			}
			defer csvW.Close()
			fmt.Fprintln(os.Stderr, "  CSV output enabled")
		}

		// Metadata
		meta = newMetadata(*outDir, *portName, *baudRate, *measRate, *antenna, *mobility, *skyvis, *weather, *anomaly, *notes)
		if err := meta.Write(); err != nil {
			log.Fatalf("Failed to write metadata: %v", err)
		}
	} else if *outFile != "" {
		// Simple raw file output (ubxcap-compatible mode)
		f, err := os.Create(*outFile)
		if err != nil {
			log.Fatalf("Failed to create %s: %v", *outFile, err)
		}
		defer func() {
			f.Sync()
			f.Close()
		}()
		reader = io.TeeReader(p, f)
		fmt.Fprintf(os.Stderr, "  Saving raw data to %s\n", *outFile)
	}

	// Signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		p.Close()
	}()

	fmt.Fprintln(os.Stderr, "Receiving... (Ctrl+C to stop)")
	fmt.Fprintln(os.Stderr)

	// Decode loop
	dec := ubx.NewDecoder(reader)
	var msgCount atomic.Int64
	status := newStatusTracker()
	gaps := newGapDetector(*measRate)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		msg, err := dec.Decode()
		if err != nil {
			break
		}
		msgCount.Add(1)

		// Track statistics
		switch m := msg.(type) {
		case *ubx.NavPVT:
			status.RecordPVT(m)
			gaps.Check(m)
			if meta != nil {
				meta.RecordStartPosition(m)
			}
		case *ubx.NavSig:
			status.RecordSig(m)
		}

		// Periodic flush + status display
		select {
		case <-ticker.C:
			if rec != nil {
				rec.Sync()
			}
			if csvW != nil {
				csvW.Flush()
			}
			status.Print(rec)
		default:
		}

		if csvW != nil {
			csvW.WriteMessage(msg)
		}

		if !*quiet {
			printMessage(msg)
		}
	}

	// Write gap log and update metadata
	if meta != nil {
		meta.GapCount = gaps.GapCount()
		meta.GapTotalMs = gaps.TotalGapMs()
		meta.Write()
		if *csvFlag {
			gaps.WriteCSV(filepath.Join(*outDir, "parsed"))
		}
	}

	fmt.Fprintf(os.Stderr, "\n\nReceived %d messages total\n", msgCount.Load())
	gaps.PrintSummary()
	if rec != nil {
		fmt.Fprintf(os.Stderr, "  Raw data: %s\n", rec.SizeString())
		if rec.WriteErrors() > 0 {
			fmt.Fprintf(os.Stderr, "  WARNING: %d write errors occurred\n", rec.WriteErrors())
		}
	}
}

func detectPort() string {
	// Linux: /dev/ttyACM*
	matches, _ := filepath.Glob("/dev/ttyACM*")
	if len(matches) > 0 {
		return matches[0]
	}
	// macOS: /dev/cu.usbmodem*
	matches, _ = filepath.Glob("/dev/cu.usbmodem*")
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}
