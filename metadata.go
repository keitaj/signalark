package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// metadata records the collection conditions for reproducibility.
type metadata struct {
	StartTime  string   `json:"startTime"`
	Port       string   `json:"port"`
	BaudRate   int      `json:"baudRate"`
	MeasRateMs int      `json:"measRateMs"`
	Messages   []string `json:"messages"`
	Location   string   `json:"location,omitempty"`
	Antenna    string   `json:"antenna,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

func writeMetadata(dir, portName string, baudRate, measRateMs int, location, antenna, notes string) error {
	m := metadata{
		StartTime:  time.Now().UTC().Format(time.RFC3339),
		Port:       portName,
		BaudRate:   baudRate,
		MeasRateMs: measRateMs,
		Messages:   []string{"NAV-PVT", "NAV-SIG", "RXM-RAWX", "MON-RF", "RXM-SFRBX"},
		Location:   location,
		Antenna:    antenna,
		Notes:      notes,
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(os.Stderr, "  Metadata saved to %s\n", path)
	return nil
}
