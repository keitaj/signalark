package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

func TestCSVWritersRoundTrip(t *testing.T) {
	dir := t.TempDir()

	writers, err := newCSVWriters(dir)
	if err != nil {
		t.Fatalf("newCSVWriters: %v", err)
	}

	// Write sample messages
	writers.WriteMessage(&ubx.NavPVT{
		ITOW: 112638005, FixType: 3, NumSV: 9,
		Lon: 1394634722, Lat: 353356739,
		HMSL: 11300, HAcc: 5029, VAcc: 8515, PDOP: 209,
	})
	writers.WriteMessage(&ubx.NavSig{
		ITOW: 112638005, NumSigs: 1,
		Signals: []ubx.NavSigSignal{
			{GnssID: 0, SvID: 1, CNO: 35, QualityInd: 7, SigFlags: 0x0029},
		},
	})
	writers.WriteMessage(&ubx.MonRF{
		Blocks: []ubx.MonRFBlock{
			{BlockID: 0, JamInd: 20, AgcCnt: 65},
		},
	})
	writers.WriteMessage(&ubx.NavSAT{
		ITOW: 112638005, NumSvs: 1,
		Svs: []ubx.NavSATSv{
			{GnssID: 0, SvID: 1, CNO: 35, Elev: 45, Azim: 120, Flags: 0x081F},
		},
	})

	writers.Flush()
	writers.Close()

	// Verify nav_pvt.csv
	assertCSVRows(t, filepath.Join(dir, "nav_pvt.csv"), 1)

	// Verify nav_sig.csv
	assertCSVRows(t, filepath.Join(dir, "nav_sig.csv"), 1)

	// Verify mon_rf.csv
	assertCSVRows(t, filepath.Join(dir, "mon_rf.csv"), 1)

	// Verify nav_sat.csv
	assertCSVRows(t, filepath.Join(dir, "nav_sat.csv"), 1)

	// Verify rxm_rawx.csv exists with header only
	assertCSVRows(t, filepath.Join(dir, "rxm_rawx.csv"), 0)

	// Verify rxm_sfrbx.csv exists with header only
	assertCSVRows(t, filepath.Join(dir, "rxm_sfrbx.csv"), 0)
}

func TestCSVNavPVTFields(t *testing.T) {
	dir := t.TempDir()
	writers, err := newCSVWriters(dir)
	if err != nil {
		t.Fatalf("newCSVWriters: %v", err)
	}

	writers.WriteMessage(&ubx.NavPVT{
		ITOW: 112638005, FixType: 3, NumSV: 18,
		Lon: 1394634722, Lat: 353356739,
		HMSL: 11300, HAcc: 1200, VAcc: 1800, PDOP: 95,
	})
	writers.Flush()
	writers.Close()

	rows := readCSVRows(t, filepath.Join(dir, "nav_pvt.csv"))
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	r := rows[0]
	// itow
	if r[0] != "112638005" {
		t.Errorf("itow: got %s, want 112638005", r[0])
	}
	// fixType
	if r[6] != "3" {
		t.Errorf("fixType: got %s, want 3", r[6])
	}
	// numSV
	if r[7] != "18" {
		t.Errorf("numSV: got %s, want 18", r[7])
	}
}

func TestCSVNavSATFields(t *testing.T) {
	dir := t.TempDir()
	writers, err := newCSVWriters(dir)
	if err != nil {
		t.Fatalf("newCSVWriters: %v", err)
	}

	writers.WriteMessage(&ubx.NavSAT{
		ITOW: 112638005, NumSvs: 2,
		Svs: []ubx.NavSATSv{
			{GnssID: 0, SvID: 1, CNO: 42, Elev: 45, Azim: 120, Flags: 0x081F},
			{GnssID: 2, SvID: 5, CNO: 38, Elev: 30, Azim: 250, Flags: 0x1214},
		},
	})
	writers.Flush()
	writers.Close()

	rows := readCSVRows(t, filepath.Join(dir, "nav_sat.csv"))
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// First satellite: GPS SV1, svUsed=1
	if rows[0][1] != "0" { // gnssId
		t.Errorf("row[0] gnssId: got %s, want 0", rows[0][1])
	}
	if rows[0][4] != "45" { // elev
		t.Errorf("row[0] elev: got %s, want 45", rows[0][4])
	}
	if rows[0][5] != "120" { // azim
		t.Errorf("row[0] azim: got %s, want 120", rows[0][5])
	}
	if rows[0][7] != "1" { // svUsed
		t.Errorf("row[0] svUsed: got %s, want 1", rows[0][7])
	}

	// Second satellite: Galileo SV5, svUsed=0
	if rows[1][1] != "2" { // gnssId
		t.Errorf("row[1] gnssId: got %s, want 2", rows[1][1])
	}
	if rows[1][7] != "0" { // svUsed
		t.Errorf("row[1] svUsed: got %s, want 0", rows[1][7])
	}
}

// assertCSVRows checks that a CSV file has the expected number of data rows (excluding header).
func assertCSVRows(t *testing.T, path string, expectedRows int) {
	t.Helper()
	rows := readCSVRows(t, path)
	if len(rows) != expectedRows {
		t.Errorf("%s: expected %d rows, got %d", filepath.Base(path), expectedRows, len(rows))
	}
}

func readCSVRows(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(records) < 1 {
		t.Fatalf("%s: expected at least header row", path)
	}
	return records[1:] // skip header
}
