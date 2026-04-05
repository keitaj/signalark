package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// sessionSummary accumulates statistics during a collection session.
type sessionSummary struct {
	totalEpochs int
	fixEpochs   int
	firstITOW   uint32
	lastITOW    uint32
	hasFirst    bool
	cnoMin      uint8
	cnoMax      uint8
	cnoSum      uint64
	cnoCount    uint64
}

// RecordPVT updates epoch and fix counters from a NAV-PVT message.
func (s *sessionSummary) RecordPVT(pvt *ubx.NavPVT) {
	s.totalEpochs++
	if pvt.FixType >= 2 {
		s.fixEpochs++
	}
	if !s.hasFirst {
		s.firstITOW = pvt.ITOW
		s.hasFirst = true
	}
	s.lastITOW = pvt.ITOW
}

// RecordSig updates CN0 statistics from a NAV-SIG message.
// Signals with CNO == 0 are skipped.
func (s *sessionSummary) RecordSig(sig *ubx.NavSig) {
	for _, ss := range sig.Signals {
		if ss.CNO == 0 {
			continue
		}
		if s.cnoCount == 0 || ss.CNO < s.cnoMin {
			s.cnoMin = ss.CNO
		}
		if ss.CNO > s.cnoMax {
			s.cnoMax = ss.CNO
		}
		s.cnoSum += uint64(ss.CNO)
		s.cnoCount++
	}
}

// FixRate returns the proportion of epochs with a valid fix (0.0–1.0).
func (s *sessionSummary) FixRate() float64 {
	if s.totalEpochs == 0 {
		return 0
	}
	return float64(s.fixEpochs) / float64(s.totalEpochs)
}

// CnoAvg returns the mean CN0 across all recorded signals.
func (s *sessionSummary) CnoAvg() float64 {
	if s.cnoCount == 0 {
		return 0
	}
	return float64(s.cnoSum) / float64(s.cnoCount)
}

// Duration computes the collection duration from the first to last ITOW.
// It handles GPS week wraparound (ITOW resets at 604800000 ms).
func (s *sessionSummary) Duration() time.Duration {
	if !s.hasFirst {
		return 0
	}
	diff := int64(s.lastITOW) - int64(s.firstITOW)
	if diff < 0 {
		// Week wraparound: 604800000 ms = 7 days in milliseconds.
		diff += 604800000
	}
	return time.Duration(diff) * time.Millisecond
}

// Print writes a human-readable session summary to stderr.
func (s *sessionSummary) Print() {
	if s.totalEpochs == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "Session summary:")
	fmt.Fprintf(os.Stderr, "  Duration:    %s\n", s.Duration())
	fmt.Fprintf(os.Stderr, "  Epochs:      %d (fix rate: %.1f%%)\n",
		s.totalEpochs, s.FixRate()*100)
	if s.cnoCount > 0 {
		fmt.Fprintf(os.Stderr, "  CN0 (dB-Hz): min=%d avg=%.1f max=%d\n",
			s.cnoMin, math.Round(s.CnoAvg()*10)/10, s.cnoMax)
	}
}
