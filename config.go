package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

func boolToRate(enabled bool) uint8 {
	if enabled {
		return 1
	}
	return 0
}

func configure(rw io.ReadWriter, measRateMs int, msgs MessageSet) {
	fmt.Fprintln(os.Stderr, "Configuring receiver...")

	// Explicitly enable or disable each message so that leftover RAM
	// configuration from a prior run does not produce unexpected output.
	b := ubx.NewCfgValset(ubx.LayerRAM).
		AddU1(ubx.KeyMsgoutNavPvtUSB, boolToRate(msgs.NavPVT)).
		AddU1(ubx.KeyMsgoutNavSigUSB, boolToRate(msgs.NavSig)).
		AddU1(ubx.KeyMsgoutMonRfUSB, boolToRate(msgs.MonRF)).
		AddU1(ubx.KeyMsgoutRxmRawxUSB, boolToRate(msgs.RxmRAWX)).
		AddU1(ubx.KeyMsgoutRxmSfrbxUSB, boolToRate(msgs.RxmSFRBX))
	b.AddU2(ubx.KeyRateMeas, uint16(measRateMs))

	if _, err := rw.Write(b.Build()); err != nil {
		log.Fatalf("Failed to send configuration: %v", err)
	}
	waitForAck(rw)
	fmt.Fprintf(os.Stderr, "  Enabled: %s\n", strings.Join(msgs.Names(), ", "))
	fmt.Fprintf(os.Stderr, "  Measurement rate: %dms (%dHz)\n", measRateMs, 1000/measRateMs)
}

// waitForAck reads UBX messages until an ACK-ACK or ACK-NAK for CFG-VALSET
// is received, or until a 2-second timeout expires.
// NOTE: On timeout or error, log.Fatal terminates the process, so the
// goroutine does not leak. If this is refactored to return an error,
// use context.Context to cancel the goroutine.
func waitForAck(r io.Reader) {
	type ackResult struct {
		ok  bool  // true = ACK-ACK, false = ACK-NAK
		err error // non-nil if decode failed
	}

	dec := ubx.NewDecoder(r)
	ch := make(chan ackResult, 1)

	go func() {
		for {
			msg, err := dec.Decode()
			if err != nil {
				ch <- ackResult{err: err}
				return
			}
			switch m := msg.(type) {
			case *ubx.AckAck:
				if m.AckedClass == ubx.ClassCFG && m.AckedID == ubx.IDCfgValset {
					ch <- ackResult{ok: true}
					return
				}
			case *ubx.AckNak:
				if m.NakedClass == ubx.ClassCFG && m.NakedID == ubx.IDCfgValset {
					ch <- ackResult{ok: false}
					return
				}
			}
		}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			log.Fatalf("Failed to read ACK for CFG-VALSET: %v", res.err)
		}
		if res.ok {
			fmt.Fprintln(os.Stderr, "  CFG-VALSET accepted (ACK-ACK)")
		} else {
			log.Fatal("CFG-VALSET rejected by receiver (ACK-NAK)")
		}
	case <-time.After(2 * time.Second):
		log.Fatal("Timeout waiting for CFG-VALSET acknowledgment")
	}
}
