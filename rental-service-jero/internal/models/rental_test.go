package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestRental_BeforeCreate_GeneratesUUID(t *testing.T) {
	r := &Rental{}
	if err := r.BeforeCreate(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be generated")
	}
}

func TestRental_BeforeCreate_PreservesExistingUUID(t *testing.T) {
	existingID := uuid.New()
	r := &Rental{ID: existingID}
	if err := r.BeforeCreate(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != existingID {
		t.Errorf("expected ID %s to be preserved, got %s", existingID, r.ID)
	}
}

func TestRental_BeforeCreate_SetsDefaultStatus(t *testing.T) {
	r := &Rental{}
	if err := r.BeforeCreate(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusActive {
		t.Errorf("expected status %s, got %s", StatusActive, r.Status)
	}
}

func TestRental_BeforeCreate_PreservesExistingStatus(t *testing.T) {
	r := &Rental{Status: StatusFinalized}
	if err := r.BeforeCreate(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusFinalized {
		t.Errorf("expected status %s, got %s", StatusFinalized, r.Status)
	}
}

func TestRental_BeforeCreate_SetsStartTime(t *testing.T) {
	r := &Rental{}
	if err := r.BeforeCreate(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusActive != "active" {
		t.Errorf("unexpected StatusActive: %s", StatusActive)
	}
	if StatusFinalized != "finalized" {
		t.Errorf("unexpected StatusFinalized: %s", StatusFinalized)
	}
	if StatusCancelled != "cancelled" {
		t.Errorf("unexpected StatusCancelled: %s", StatusCancelled)
	}
}
