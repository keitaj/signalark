package main

import (
	"reflect"
	"testing"
)

func TestParseMessagesDefault(t *testing.T) {
	ms := parseMessages("nav-pvt,nav-sat,nav-sig,mon-rf,rxm-rawx,rxm-sfrbx")
	if !ms.NavPVT || !ms.NavSAT || !ms.NavSig || !ms.MonRF || !ms.RxmRAWX || !ms.RxmSFRBX {
		t.Error("expected all messages enabled")
	}
}

func TestParseMessagesSubset(t *testing.T) {
	ms := parseMessages("nav-pvt,nav-sig")
	if !ms.NavPVT {
		t.Error("expected NavPVT enabled")
	}
	if !ms.NavSig {
		t.Error("expected NavSig enabled")
	}
	if ms.MonRF {
		t.Error("expected MonRF disabled")
	}
	if ms.RxmRAWX {
		t.Error("expected RxmRAWX disabled")
	}
}

func TestParseMessagesCaseInsensitive(t *testing.T) {
	ms := parseMessages("NAV-PVT,Nav-Sig")
	if !ms.NavPVT || !ms.NavSig {
		t.Error("expected case-insensitive parsing")
	}
}

func TestParseMessagesWithSpaces(t *testing.T) {
	ms := parseMessages(" nav-pvt , nav-sig ")
	if !ms.NavPVT || !ms.NavSig {
		t.Error("expected trimmed parsing")
	}
}

func TestParseMessagesTrailingComma(t *testing.T) {
	ms := parseMessages("nav-pvt,")
	if !ms.NavPVT {
		t.Error("expected NavPVT enabled")
	}
}

func TestMessageSetNames(t *testing.T) {
	ms := MessageSet{NavPVT: true, NavSig: true, MonRF: true}
	want := []string{"NAV-PVT", "NAV-SIG", "MON-RF"}
	got := ms.Names()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

func TestMessageSetNamesAll(t *testing.T) {
	ms := MessageSet{NavPVT: true, NavSAT: true, NavSig: true, MonRF: true, RxmRAWX: true, RxmSFRBX: true}
	got := ms.Names()
	if len(got) != 6 {
		t.Errorf("expected 6 names, got %d: %v", len(got), got)
	}
}
