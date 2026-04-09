package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

func configure(w io.Writer, measRateMs int) {
	fmt.Fprintln(os.Stderr, "Configuring receiver...")

	frame := ubx.NewCfgValset(ubx.LayerRAM).
		AddU1(ubx.KeyMsgoutNavPvtUSB, 1).
		AddU1(ubx.KeyMsgoutNavSigUSB, 1).
		AddU1(ubx.KeyMsgoutRxmRawxUSB, 1).
		AddU1(ubx.KeyMsgoutMonRfUSB, 1).
		AddU1(ubx.KeyMsgoutRxmSfrbxUSB, 1).
		AddU1(ubx.KeyItfmEnable, 1).
		AddU1(ubx.KeyItfmBBThreshold, 3).
		AddU1(ubx.KeyItfmCWThreshold, 15).
		AddU1(ubx.KeyItfmAntSetting, 2).
		AddU2(ubx.KeyRateMeas, uint16(measRateMs)).
		Build()

	if _, err := w.Write(frame); err != nil {
		log.Fatalf("Failed to send configuration: %v", err)
	}

	fmt.Fprintf(os.Stderr, "  NAV-PVT, NAV-SIG, RXM-RAWX, MON-RF, RXM-SFRBX enabled\n")
	fmt.Fprintf(os.Stderr, "  ITFM interference monitor enabled (active antenna)\n")
	fmt.Fprintf(os.Stderr, "  Measurement rate: %dms (%dHz)\n", measRateMs, 1000/measRateMs)
}
