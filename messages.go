package main

import (
	"log"
	"strings"
)

// MessageSet tracks which UBX messages are enabled.
type MessageSet struct {
	NavPVT   bool
	NavSAT   bool
	NavSig   bool
	MonRF    bool
	RxmRAWX  bool
	RxmSFRBX bool
}

// Names returns the enabled message names for display/metadata.
func (m MessageSet) Names() []string {
	var names []string
	if m.NavPVT {
		names = append(names, "NAV-PVT")
	}
	if m.NavSAT {
		names = append(names, "NAV-SAT")
	}
	if m.NavSig {
		names = append(names, "NAV-SIG")
	}
	if m.MonRF {
		names = append(names, "MON-RF")
	}
	if m.RxmRAWX {
		names = append(names, "RXM-RAWX")
	}
	if m.RxmSFRBX {
		names = append(names, "RXM-SFRBX")
	}
	return names
}

var validMessages = map[string]bool{
	"nav-pvt": true, "nav-sat": true, "nav-sig": true, "mon-rf": true,
	"rxm-rawx": true, "rxm-sfrbx": true,
}

func parseMessages(s string) MessageSet {
	s = strings.TrimSpace(s)
	if s == "" {
		log.Fatal("-messages must not be empty (valid: nav-pvt, nav-sat, nav-sig, mon-rf, rxm-rawx, rxm-sfrbx)")
	}
	var ms MessageSet
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		// validMessages and switch must stay in sync when adding new message types.
		if !validMessages[name] {
			log.Fatalf("Unknown message type: %q (valid: nav-pvt, nav-sat, nav-sig, mon-rf, rxm-rawx, rxm-sfrbx)", name)
		}
		switch name {
		case "nav-pvt":
			ms.NavPVT = true
		case "nav-sat":
			ms.NavSAT = true
		case "nav-sig":
			ms.NavSig = true
		case "mon-rf":
			ms.MonRF = true
		case "rxm-rawx":
			ms.RxmRAWX = true
		case "rxm-sfrbx":
			ms.RxmSFRBX = true
		}
	}
	return ms
}
