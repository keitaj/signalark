package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// diagItfm sends each CFG-ITFM-* key as a separate single-key CFG-VALSET
// and reports the per-key ACK/NAK result. Used to bisect which keys F9P
// firmware rejects when batched. Throwaway diagnostic; do not merge.
func diagItfm(rw io.ReadWriter) {
	cases := []struct {
		name string
		key  uint32
		val  uint8
	}{
		{"CFG-ITFM-ENABLE       (0x1041000d)", ubx.KeyItfmEnable, 1},
		{"CFG-ITFM-ANTSETTING   (0x20410010)", ubx.KeyItfmAntSetting, 2},
		{"CFG-ITFM-BBTHRESHOLD  (0x20410001)", ubx.KeyItfmBBThreshold, 3},
		{"CFG-ITFM-CWTHRESHOLD  (0x20410002)", ubx.KeyItfmCWThreshold, 15},
	}

	fmt.Fprintln(os.Stderr, "ITFM bisect diagnostic — sending each key as its own CFG-VALSET to RAM:")
	for _, c := range cases {
		msg := ubx.NewCfgValset(ubx.LayerRAM).AddU1(c.key, c.val).Build()
		if _, err := rw.Write(msg); err != nil {
			fmt.Fprintf(os.Stderr, "  %s val=%d → WRITE ERROR: %v\n", c.name, c.val, err)
			continue
		}
		result := waitForAckNonFatal(rw, 2*time.Second)
		fmt.Fprintf(os.Stderr, "  %s val=%d → %s\n", c.name, c.val, result)
	}
}

// waitForAckNonFatal reads UBX messages until an ACK-ACK or ACK-NAK for
// CFG-VALSET is received, or until timeout. Unlike waitForAck, this does
// NOT exit on NAK / timeout — it returns a short status string instead.
func waitForAckNonFatal(r io.Reader, timeout time.Duration) string {
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)

	go func() {
		dec := ubx.NewDecoder(r)
		for {
			msg, err := dec.Decode()
			if err != nil {
				ch <- result{err: err}
				return
			}
			switch m := msg.(type) {
			case *ubx.AckAck:
				if m.AckedClass == ubx.ClassCFG && m.AckedID == ubx.IDCfgValset {
					ch <- result{s: "ACK-ACK"}
					return
				}
			case *ubx.AckNak:
				if m.NakedClass == ubx.ClassCFG && m.NakedID == ubx.IDCfgValset {
					ch <- result{s: "ACK-NAK"}
					return
				}
			}
		}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return fmt.Sprintf("DECODE ERROR: %v", r.err)
		}
		return r.s
	case <-time.After(timeout):
		return "TIMEOUT"
	}
}
