package messaging

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

type mockBikeRepo struct {
	upsertCalled bool
	deleteCalled bool
	lastID       uuid.UUID
	err          error
}

func (m *mockBikeRepo) Upsert(_ context.Context, bikeID uuid.UUID) error {
	m.upsertCalled = true
	m.lastID = bikeID
	return m.err
}

func (m *mockBikeRepo) Delete(_ context.Context, bikeID uuid.UUID) error {
	m.deleteCalled = true
	m.lastID = bikeID
	return m.err
}

func (m *mockBikeRepo) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

func TestHandleEvent_Created(t *testing.T) {
	repo := &mockBikeRepo{}
	consumer := &Consumer{bikeRepo: repo}

	bikeID := uuid.New()
	event := BikeLifecycleEvent{BikeID: bikeID.String(), Action: "CREATED"}
	body, _ := json.Marshal(event)

	err := consumer.HandleEvent(context.Background(), body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.upsertCalled {
		t.Error("expected Upsert to be called")
	}
	if repo.lastID != bikeID {
		t.Errorf("expected bike ID %s, got %s", bikeID, repo.lastID)
	}
}

func TestHandleEvent_Deleted(t *testing.T) {
	repo := &mockBikeRepo{}
	consumer := &Consumer{bikeRepo: repo}

	bikeID := uuid.New()
	event := BikeLifecycleEvent{BikeID: bikeID.String(), Action: "DELETED"}
	body, _ := json.Marshal(event)

	err := consumer.HandleEvent(context.Background(), body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleteCalled {
		t.Error("expected Delete to be called")
	}
}

func TestHandleEvent_UnknownAction(t *testing.T) {
	repo := &mockBikeRepo{}
	consumer := &Consumer{bikeRepo: repo}

	event := BikeLifecycleEvent{BikeID: uuid.New().String(), Action: "UNKNOWN"}
	body, _ := json.Marshal(event)

	err := consumer.HandleEvent(context.Background(), body)
	if err != nil {
		t.Fatalf("unexpected error for unknown action: %v", err)
	}
	if repo.upsertCalled || repo.deleteCalled {
		t.Error("no repo method should be called for unknown action")
	}
}

func TestHandleEvent_MalformedJSON(t *testing.T) {
	consumer := &Consumer{bikeRepo: &mockBikeRepo{}}
	err := consumer.HandleEvent(context.Background(), []byte("not json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestHandleEvent_InvalidBikeID(t *testing.T) {
	consumer := &Consumer{bikeRepo: &mockBikeRepo{}}
	event := BikeLifecycleEvent{BikeID: "not-a-uuid", Action: "CREATED"}
	body, _ := json.Marshal(event)

	err := consumer.HandleEvent(context.Background(), body)
	if err == nil {
		t.Error("expected error for invalid bike_id")
	}
}

func TestHandleEvent_RepoError(t *testing.T) {
	repo := &mockBikeRepo{err: context.DeadlineExceeded}
	consumer := &Consumer{bikeRepo: repo}

	event := BikeLifecycleEvent{BikeID: uuid.New().String(), Action: "CREATED"}
	body, _ := json.Marshal(event)

	err := consumer.HandleEvent(context.Background(), body)
	if err == nil {
		t.Error("expected error when repo fails")
	}
}
