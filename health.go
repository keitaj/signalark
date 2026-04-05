package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// runHealthCheck performs startup checks after receiver configuration:
// 1. Waits for ACK-ACK confirming CFG-VALSET was accepted.
// 2. Waits for a NAV-PVT with fixType >= 2 (at least a 2D fix).
// 3. Checks the first MON-RF message for antenna status warnings.
//
// It reads directly from the serial port before the TeeReader is set up.
// Warnings are logged to stderr but do not prevent collection from proceeding.
func runHealthCheck(port io.ReadWriter, timeout time.Duration) error {
	fmt.Fprintln(os.Stderr, "Running startup health check...")

	deadline := time.Now().Add(timeout)
	ackDeadline := time.Now().Add(3 * time.Second)

	dec := ubx.NewDecoder(port)

	var gotACK bool
	var gotFix bool
	var gotMonRF bool

	for time.Now().Before(deadline) {
		msg, err := dec.Decode()
		if err != nil {
			return fmt.Errorf("health check decode error: %w", err)
		}

		switch m := msg.(type) {
		case *ubx.AckAck:
			if !gotACK {
				fmt.Fprintf(os.Stderr, "  ACK-ACK received: %s accepted\n", m.AckedClassID())
				gotACK = true
			}

		case *ubx.AckNak:
			if !gotACK {
				fmt.Fprintf(os.Stderr, "  WARNING: ACK-NAK received: %s rejected\n", m.NakedClassID())
				gotACK = true // treat as resolved, continue
			}

		case *ubx.NavPVT:
			if !gotFix {
				if m.FixType >= 2 {
					fmt.Fprintf(os.Stderr, "  Fix acquired: %s (sats=%d)\n", fixString(m), m.NumSV)
					gotFix = true
				} else {
					fmt.Fprintf(os.Stderr, "  Waiting for fix... (sats=%d, fixType=%d)\r", m.NumSV, m.FixType)
				}
			}

		case *ubx.MonRF:
			if !gotMonRF {
				gotMonRF = true
				for _, b := range m.Blocks {
					switch b.AntStatus {
					case 0:
						fmt.Fprintln(os.Stderr, "  WARNING: Antenna status: INIT (no antenna supervisor)")
					case 3:
						fmt.Fprintln(os.Stderr, "  WARNING: Antenna status: SHORT detected")
					default:
						fmt.Fprintf(os.Stderr, "  Antenna status: OK (block=%d, antStatus=%d)\n", b.BlockID, b.AntStatus)
					}
				}
			}
		}

		// Check ACK timeout separately (3s)
		if !gotACK && time.Now().After(ackDeadline) {
			fmt.Fprintln(os.Stderr, "  WARNING: No ACK response within 3s, continuing")
			gotACK = true
		}

		// All checks done
		if gotACK && gotFix && gotMonRF {
			break
		}
	}

	// Report any timeouts
	if !gotACK {
		fmt.Fprintln(os.Stderr, "  WARNING: No ACK response within 3s, continuing")
	}
	if !gotFix {
		fmt.Fprintln(os.Stderr, "\n  WARNING: No fix acquired within timeout, starting collection anyway")
	}
	if !gotMonRF {
		fmt.Fprintln(os.Stderr, "  WARNING: No MON-RF received within timeout")
	}

	fmt.Fprintln(os.Stderr, "Health check complete.")
	return nil
}
