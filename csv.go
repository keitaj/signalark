package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

// csvWriters holds all CSV writers for structured data output.
type csvWriters struct {
	pvt  *navPVTCSV
	sig  *navSigCSV
	rf   *monRFCSV
	rawx *rxmRAWXCSV
}

func newCSVWriters(dir string) (*csvWriters, error) {
	pvt, err := newNavPVTCSV(dir)
	if err != nil {
		return nil, err
	}
	sig, err := newNavSigCSV(dir)
	if err != nil {
		pvt.Close()
		return nil, err
	}
	rf, err := newMonRFCSV(dir)
	if err != nil {
		pvt.Close()
		sig.Close()
		return nil, err
	}
	rawx, err := newRxmRAWXCSV(dir)
	if err != nil {
		pvt.Close()
		sig.Close()
		rf.Close()
		return nil, err
	}
	return &csvWriters{pvt: pvt, sig: sig, rf: rf, rawx: rawx}, nil
}

func (c *csvWriters) WriteMessage(msg ubx.Message) {
	switch m := msg.(type) {
	case *ubx.NavPVT:
		c.pvt.Write(m)
	case *ubx.NavSig:
		c.sig.Write(m)
	case *ubx.MonRF:
		c.rf.Write(m)
	case *ubx.RxmRAWX:
		c.rawx.Write(m)
	}
}

func (c *csvWriters) Flush() {
	c.pvt.w.Flush()
	c.sig.w.Flush()
	c.rf.w.Flush()
	c.rawx.w.Flush()
}

func (c *csvWriters) Close() {
	c.pvt.Close()
	c.sig.Close()
	c.rf.Close()
	c.rawx.Close()
}

// --- NAV-PVT CSV ---

type navPVTCSV struct {
	w *csv.Writer
	f *os.File
}

func newNavPVTCSV(dir string) (*navPVTCSV, error) {
	f, err := createCSV(dir, "nav_pvt.csv")
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Write([]string{"itow", "lat", "lon", "alt", "hAcc", "vAcc", "fixType", "numSV", "pDOP"})
	return &navPVTCSV{w: w, f: f}, nil
}

func (c *navPVTCSV) Write(m *ubx.NavPVT) {
	c.w.Write([]string{
		strconv.FormatUint(uint64(m.ITOW), 10),
		strconv.FormatFloat(m.LatDeg(), 'f', 7, 64),
		strconv.FormatFloat(m.LonDeg(), 'f', 7, 64),
		strconv.FormatFloat(m.HMSLM(), 'f', 3, 64),
		strconv.FormatFloat(m.HAccM(), 'f', 3, 64),
		strconv.FormatFloat(m.VAccM(), 'f', 3, 64),
		strconv.FormatUint(uint64(m.FixType), 10),
		strconv.FormatUint(uint64(m.NumSV), 10),
		strconv.FormatFloat(m.PDOPVal(), 'f', 2, 64),
	})
}

func (c *navPVTCSV) Close() error {
	c.w.Flush()
	return c.f.Close()
}

// --- NAV-SIG CSV ---

type navSigCSV struct {
	w *csv.Writer
	f *os.File
}

func newNavSigCSV(dir string) (*navSigCSV, error) {
	f, err := createCSV(dir, "nav_sig.csv")
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Write([]string{"itow", "gnssId", "svId", "sigId", "cno", "qualityInd", "sigFlags"})
	return &navSigCSV{w: w, f: f}, nil
}

func (c *navSigCSV) Write(m *ubx.NavSig) {
	itow := strconv.FormatUint(uint64(m.ITOW), 10)
	for _, s := range m.Signals {
		c.w.Write([]string{
			itow,
			strconv.FormatUint(uint64(s.GnssID), 10),
			strconv.FormatUint(uint64(s.SvID), 10),
			strconv.FormatUint(uint64(s.SigID), 10),
			strconv.FormatUint(uint64(s.CNO), 10),
			strconv.FormatUint(uint64(s.QualityInd), 10),
			fmt.Sprintf("0x%04X", s.SigFlags),
		})
	}
}

func (c *navSigCSV) Close() error {
	c.w.Flush()
	return c.f.Close()
}

// --- MON-RF CSV ---

type monRFCSV struct {
	w *csv.Writer
	f *os.File
}

func newMonRFCSV(dir string) (*monRFCSV, error) {
	f, err := createCSV(dir, "mon_rf.csv")
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Write([]string{"timestampMs", "blockId", "jammingState", "jamInd", "agcCnt"})
	return &monRFCSV{w: w, f: f}, nil
}

func (c *monRFCSV) Write(m *ubx.MonRF) {
	ts := strconv.FormatInt(nowUnixMs(), 10)
	for _, b := range m.Blocks {
		c.w.Write([]string{
			ts,
			strconv.FormatUint(uint64(b.BlockID), 10),
			b.JammingState(),
			strconv.FormatUint(uint64(b.JamInd), 10),
			strconv.FormatUint(uint64(b.AgcCnt), 10),
		})
	}
}

func (c *monRFCSV) Close() error {
	c.w.Flush()
	return c.f.Close()
}

// --- RXM-RAWX CSV ---

type rxmRAWXCSV struct {
	w *csv.Writer
	f *os.File
}

func newRxmRAWXCSV(dir string) (*rxmRAWXCSV, error) {
	f, err := createCSV(dir, "rxm_rawx.csv")
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Write([]string{"rcvTow", "week", "gnssId", "svId", "sigId", "prMes", "cpMes", "doMes", "cno", "trkStat"})
	return &rxmRAWXCSV{w: w, f: f}, nil
}

func (c *rxmRAWXCSV) Write(m *ubx.RxmRAWX) {
	tow := strconv.FormatFloat(m.RcvTow, 'f', 6, 64)
	week := strconv.FormatUint(uint64(m.Week), 10)
	for _, meas := range m.Meas {
		c.w.Write([]string{
			tow,
			week,
			strconv.FormatUint(uint64(meas.GnssID), 10),
			strconv.FormatUint(uint64(meas.SvID), 10),
			strconv.FormatUint(uint64(meas.SigID), 10),
			strconv.FormatFloat(meas.PrMes, 'f', 3, 64),
			strconv.FormatFloat(meas.CpMes, 'f', 3, 64),
			strconv.FormatFloat(float64(meas.DoMes), 'f', 3, 64),
			strconv.FormatUint(uint64(meas.CNO), 10),
			strconv.FormatUint(uint64(meas.TrkStat), 10),
		})
	}
}

func (c *rxmRAWXCSV) Close() error {
	c.w.Flush()
	return c.f.Close()
}

// --- helpers ---

func createCSV(dir, name string) (*os.File, error) {
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	return f, nil
}
