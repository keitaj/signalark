package main

import (
	"testing"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

func TestGapDetectorNoGaps(t *testing.T) {
	g := newGapDetector(1000)

	// Simulate 5 consecutive 1Hz epochs
	for i := uint32(0); i < 5; i++ {
		g.Check(&ubx.NavPVT{ITOW: 100000 + i*1000})
	}

	if g.GapCount() != 0 {
		t.Errorf("expected 0 gaps, got %d", g.GapCount())
	}
	if g.TotalGapMs() != 0 {
		t.Errorf("expected 0 total gap ms, got %d", g.TotalGapMs())
	}
}

func TestGapDetectorDetectsGap(t *testing.T) {
	g := newGapDetector(1000)

	g.Check(&ubx.NavPVT{ITOW: 100000})
	g.Check(&ubx.NavPVT{ITOW: 101000}) // normal
	g.Check(&ubx.NavPVT{ITOW: 105000}) // 4s gap

	if g.GapCount() != 1 {
		t.Fatalf("expected 1 gap, got %d", g.GapCount())
	}
	if g.TotalGapMs() != 4000 {
		t.Errorf("expected 4000ms total gap, got %d", g.TotalGapMs())
	}
}

func TestGapDetectorTolerance(t *testing.T) {
	g := newGapDetector(1000)

	g.Check(&ubx.NavPVT{ITOW: 100000})
	g.Check(&ubx.NavPVT{ITOW: 101400}) // 1.4s — within 1.5x tolerance

	if g.GapCount() != 0 {
		t.Errorf("expected 0 gaps for 1.4s delta at 1Hz, got %d", g.GapCount())
	}
}

func TestGapDetectorWeekWraparound(t *testing.T) {
	g := newGapDetector(1000)

	g.Check(&ubx.NavPVT{ITOW: 604799000}) // end of GPS week
	g.Check(&ubx.NavPVT{ITOW: 1000})      // start of next week

	// Should be 2s gap (604799→604800→0→1), within tolerance? 2000ms > 1500ms tolerance
	if g.GapCount() != 1 {
		t.Errorf("expected 1 gap for week wraparound, got %d", g.GapCount())
	}
}

func TestItowDelta(t *testing.T) {
	tests := []struct {
		prev, curr uint32
		want       uint32
	}{
		{100000, 101000, 1000},
		{100000, 100000, 0},
		{604799000, 1000, 2000}, // week wraparound
	}
	for _, tt := range tests {
		got := itowDelta(tt.prev, tt.curr)
		if got != tt.want {
			t.Errorf("itowDelta(%d, %d) = %d, want %d", tt.prev, tt.curr, got, tt.want)
		}
	}
}
