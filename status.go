package main

import (
	"fmt"
	"os"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// statusTracker tracks running statistics for periodic status display.
type statusTracker struct {
	start     time.Time
	epochs    int64
	fixEpochs int64
	cnoSum    uint64
	cnoCount  uint64
	lastCnoAvg float64
}

func newStatusTracker() *statusTracker {
	return &statusTracker{start: time.Now()}
}

func (s *statusTracker) RecordPVT(pvt *ubx.NavPVT) {
	s.epochs++
	if pvt.FixType >= 2 {
		s.fixEpochs++
	}
}

func (s *statusTracker) RecordSig(sig *ubx.NavSig) {
	for _, ss := range sig.Signals {
		if ss.CNO > 0 {
			s.cnoSum += uint64(ss.CNO)
			s.cnoCount++
		}
	}
	if s.cnoCount > 0 {
		s.lastCnoAvg = float64(s.cnoSum) / float64(s.cnoCount)
	}
}

// Print outputs a single-line status to stderr using \r for in-place update.
func (s *statusTracker) Print(rec *recorder) {
	elapsed := time.Since(s.start).Truncate(time.Second)
	fixPct := 0.0
	if s.epochs > 0 {
		fixPct = float64(s.fixEpochs) / float64(s.epochs) * 100
	}

	sizeStr := ""
	if rec != nil {
		sizeStr = fmt.Sprintf(" | raw: %s", rec.SizeString())
	}

	fmt.Fprintf(os.Stderr, "\r  [%s] %d epochs, fix=%.1f%%, CN0=%.1f%s    ",
		elapsed, s.epochs, fixPct, s.lastCnoAvg, sizeStr)
}
