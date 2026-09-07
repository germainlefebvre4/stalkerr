package models

import (
	"reflect"
	"testing"
)

func TestStringList_ValueScanRoundTrip(t *testing.T) {
	original := StringList{"ACTION-FR", "ANIMATION"}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	var scanned StringList
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if !reflect.DeepEqual(original, scanned) {
		t.Errorf("expected %v, got %v", original, scanned)
	}
}

func TestStringList_ValueScanEmpty(t *testing.T) {
	original := StringList{}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	var scanned StringList
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if !reflect.DeepEqual(original, scanned) {
		t.Errorf("expected %v, got %v", original, scanned)
	}
}

func TestStringList_ValueScanNil(t *testing.T) {
	var original StringList

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	if value != nil {
		t.Errorf("expected nil value for nil StringList, got %v", value)
	}

	var scanned StringList
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if scanned != nil {
		t.Errorf("expected nil after scanning nil, got %v", scanned)
	}
}

func TestStringList_ScanFromString(t *testing.T) {
	var scanned StringList
	if err := scanned.Scan(`["ACTION-FR","ANIMATION"]`); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	expected := StringList{"ACTION-FR", "ANIMATION"}
	if !reflect.DeepEqual(expected, scanned) {
		t.Errorf("expected %v, got %v", expected, scanned)
	}
}
