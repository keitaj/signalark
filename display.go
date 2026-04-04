package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/keitaj/go-ubx/pkg/ubx"
)

func printMessage(msg ubx.Message) {
	switch m := msg.(type) {
	case *ubx.NavPVT:
		fmt.Printf("[NAV-PVT] %04d-%02d-%02d %02d:%02d:%02d | fix=%-14s | sats=%2d | %.7f, %.7f | alt=%.1fm | hAcc=%.2fm | pDOP=%.2f\n",
			m.Year, m.Month, m.Day, m.Hour, m.Min, m.Sec,
			fixString(m), m.NumSV,
			m.LatDeg(), m.LonDeg(), m.HMSLM(),
			m.HAccM(), m.PDOPVal())

	case *ubx.NavSig:
		fmt.Printf("[NAV-SIG] %d signals\n", m.NumSigs)
		for i, s := range m.Signals {
			health := "?"
			switch s.Health() {
			case 1:
				health = "OK"
			case 2:
				health = "NG"
			}
			fmt.Printf("  [%2d] %s SV%3d sig=%d | CN0=%2d dB-Hz | quality=%d (%s) | health=%s | prUsed=%v crUsed=%v\n",
				i, gnssName(s.GnssID), s.SvID, s.SigID, s.CNO, s.QualityInd, s.QualityStr(), health, s.PrUsed(), s.CrUsed())
		}

	case *ubx.RxmRAWX:
		fmt.Printf("[RXM-RAWX] week=%d tow=%.3fs | %d measurements\n",
			m.Week, m.RcvTow, m.NumMeas)
		for i, meas := range m.Meas {
			pr := "-"
			if meas.PrValid() {
				pr = fmt.Sprintf("%.3f", meas.PrMes)
			}
			fmt.Printf("  [%2d] %s SV%3d sig=%d | CN0=%2d dB-Hz | PR=%s m\n",
				i, gnssName(meas.GnssID), meas.SvID, meas.SigID, meas.CNO, pr)
		}

	case *ubx.MonRF:
		for _, b := range m.Blocks {
			fmt.Printf("[MON-RF] block=%d | jamming=%-8s | jamInd=%3d | agc=%5d | ant=%d/%d\n",
				b.BlockID, b.JammingState(), b.JamInd, b.AgcCnt, b.AntStatus, b.AntPower)
		}

	case *ubx.RxmSFRBX:
		n := len(m.Dwrd)
		if n > 3 {
			n = 3
		}
		words := make([]string, n)
		for i := 0; i < n; i++ {
			words[i] = fmt.Sprintf("%08X", m.Dwrd[i])
		}
		fmt.Printf("[RXM-SFRBX] %s SV%d | %d words | %s...\n",
			gnssName(m.GnssID), m.SvID, m.NumWords, strings.Join(words, " "))

	case *ubx.AckAck:
		fmt.Printf("[ACK-ACK] %s accepted\n", m.AckedClassID())

	case *ubx.AckNak:
		fmt.Printf("[ACK-NAK] %s rejected\n", m.NakedClassID())

	case *ubx.CfgValget:
		fmt.Printf("[CFG-VALGET] layer=%d | %d key-value pairs\n", m.Layer, len(m.KeyVals))
		for _, kv := range m.KeyVals {
			fmt.Printf("  key=0x%08X val=%s\n", kv.Key, hex.EncodeToString(kv.Val))
		}

	case *ubx.RawMessage:
		fmt.Printf("[%s] %d bytes payload\n", m.GetClassID(), len(m.Payload))
	}
}

func fixString(m *ubx.NavPVT) string {
	s := "No fix"
	switch m.FixType {
	case 2:
		s = "2D"
	case 3:
		s = "3D"
	case 4:
		s = "GNSS+DR"
	case 5:
		s = "Time only"
	}
	if m.Flags&0x01 != 0 {
		switch (m.Flags >> 6) & 0x03 {
		case 1:
			s += "/Float"
		case 2:
			s += "/Fixed"
		}
	}
	return s
}

func gnssName(id uint8) string {
	switch id {
	case ubx.GnssIDGPS:
		return "GPS    "
	case ubx.GnssIDSBAS:
		return "SBAS   "
	case ubx.GnssIDGalileo:
		return "Galileo"
	case ubx.GnssIDBeiDou:
		return "BeiDou "
	case ubx.GnssIDQZSS:
		return "QZSS   "
	case ubx.GnssIDGLONASS:
		return "GLONASS"
	default:
		return fmt.Sprintf("GNSS(%d)", id)
	}
}
