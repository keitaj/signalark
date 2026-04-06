package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

const gpsWeekMs = 604800000 // milliseconds in a GPS week

// gapRecord represents a single detected data gap.
type gapRecord struct {
	ITOW     uint32 // ITOW after the gap
	PrevITOW uint32 // ITOW before the gap
	GapMs    uint32 // gap duration in ms
}

// gapDetector checks NAV-PVT ITOW continuity and records gaps.
type gapDetector struct {
	expectedMs uint32
	tolerance  uint32 // expectedMs * 1.5
	lastITOW   uint32
	hasFirst   bool
	gaps       []gapRecord
}

func newGapDetector(measRateMs int) *gapDetector {
	expected := uint32(measRateMs)
	return &gapDetector{
		expectedMs: expected,
		tolerance:  expected + expected/2, // 1.5x
	}
}

// Check checks ITOW continuity. Call on every NAV-PVT.
func (g *gapDetector) Check(pvt *ubx.NavPVT) {
	if !g.hasFirst {
		g.lastITOW = pvt.ITOW
		g.hasFirst = true
		return
	}

	delta := itowDelta(g.lastITOW, pvt.ITOW)
	if delta > g.tolerance {
		rec := gapRecord{
			ITOW:     pvt.ITOW,
			PrevITOW: g.lastITOW,
			GapMs:    delta,
		}
		g.gaps = append(g.gaps, rec)
		fmt.Fprintf(os.Stderr, "\nWARNING: data gap detected: %dms (ITOW %d → %d)\n", delta, g.lastITOW, pvt.ITOW)
	}
	g.lastITOW = pvt.ITOW
}

// GapCount returns the total number of gaps detected.
func (g *gapDetector) GapCount() int { return len(g.gaps) }

// TotalGapMs returns the total gap duration in milliseconds.
func (g *gapDetector) TotalGapMs() uint64 {
	var total uint64
	for _, r := range g.gaps {
		total += uint64(r.GapMs)
	}
	return total
}

// WriteCSV writes the gap log to a CSV file in the given directory.
// Only creates the file if there are gaps to record.
func (g *gapDetector) WriteCSV(dir string) error {
	if len(g.gaps) == 0 {
		return nil
	}

	path := filepath.Join(dir, "gaps.csv")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(bufio.NewWriter(f))
	w.Write([]string{"prevITOW", "itow", "gapMs"})
	for _, r := range g.gaps {
		w.Write([]string{
			strconv.FormatUint(uint64(r.PrevITOW), 10),
			strconv.FormatUint(uint64(r.ITOW), 10),
			strconv.FormatUint(uint64(r.GapMs), 10),
		})
	}
	w.Flush()

	fmt.Fprintf(os.Stderr, "  Gap log saved to %s\n", path)
	return nil
}

// PrintSummary prints gap summary to stderr if any gaps were detected.
func (g *gapDetector) PrintSummary() {
	if len(g.gaps) == 0 {
		return
	}
	total := time.Duration(g.TotalGapMs()) * time.Millisecond
	fmt.Fprintf(os.Stderr, "  Data gaps: %d (total %s)\n", len(g.gaps), total)
}

// itowDelta computes the difference between two ITOWs, handling GPS week wraparound.
func itowDelta(prev, curr uint32) uint32 {
	if curr >= prev {
		return curr - prev
	}
	// Week wraparound
	return (gpsWeekMs - prev) + curr
}
