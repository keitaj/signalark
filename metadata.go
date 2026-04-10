package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// metadata records the collection conditions for reproducibility.
type metadata struct {
	StartTime  string   `json:"startTime"`
	Port       string   `json:"port"`
	BaudRate   int      `json:"baudRate"`
	MeasRateMs int      `json:"measRateMs"`
	Messages   []string `json:"messages"`
	Antenna    string   `json:"antenna,omitempty"`
	Mobility   string   `json:"mobility,omitempty"`
	Skyvis     string   `json:"skyvis,omitempty"`
	Weather    string   `json:"weather,omitempty"`
	Anomaly    string   `json:"anomaly,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	StartLat   float64  `json:"startLat,omitempty"`
	StartLon   float64  `json:"startLon,omitempty"`
	GapCount   int      `json:"gapCount,omitempty"`
	GapTotalMs uint64   `json:"gapTotalMs,omitempty"`

	dir  string
	once sync.Once
}

func newMetadata(dir, portName string, baudRate, measRateMs int, msgs []string, antenna, mobility, skyvis, weather, anomaly, notes string) *metadata {
	return &metadata{
		StartTime:  time.Now().UTC().Format(time.RFC3339),
		Port:       portName,
		BaudRate:   baudRate,
		MeasRateMs: measRateMs,
		Messages:   msgs,
		Antenna:    antenna,
		Mobility:   mobility,
		Skyvis:     skyvis,
		Weather:    weather,
		Anomaly:    anomaly,
		Notes:      notes,
		dir:        dir,
	}
}

// RecordStartPosition records the lat/lon from the first valid NAV-PVT fix.
// Safe to call from the decode loop on every NAV-PVT; only the first fix is recorded.
func (m *metadata) RecordStartPosition(pvt *ubx.NavPVT) {
	if pvt.FixType < 2 {
		return // no fix yet
	}
	m.once.Do(func() {
		m.StartLat = pvt.LatDeg()
		m.StartLon = pvt.LonDeg()
	})
}

func (m *metadata) Write() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(m.dir, "metadata.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(os.Stderr, "  Metadata saved to %s\n", path)
	return nil
}
