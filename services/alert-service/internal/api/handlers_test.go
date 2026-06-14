package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

type fakeAlertStore struct {
	listAlerts  func(ctx context.Context, status, severity string, limit, offset int) ([]store.Alert, error)
	getAlert    func(ctx context.Context, alertID string) (store.Alert, error)
	acknowledge func(ctx context.Context, alertID, acknowledgedBy string) (store.Alert, error)
}

func (f *fakeAlertStore) ListAlerts(ctx context.Context, status, severity string, limit, offset int) ([]store.Alert, error) {
	if f.listAlerts != nil {
		return f.listAlerts(ctx, status, severity, limit, offset)
	}
	return nil, nil
}

func (f *fakeAlertStore) GetAlert(ctx context.Context, alertID string) (store.Alert, error) {
	if f.getAlert != nil {
		return f.getAlert(ctx, alertID)
	}
	return store.Alert{}, nil
}

func (f *fakeAlertStore) Acknowledge(ctx context.Context, alertID, acknowledgedBy string) (store.Alert, error) {
	if f.acknowledge != nil {
		return f.acknowledge(ctx, alertID, acknowledgedBy)
	}
	return store.Alert{}, nil
}

type fakeBroadcaster struct {
	lastType  string
	lastAlert store.Alert
}

func (f *fakeBroadcaster) Broadcast(msgType string, alert store.Alert) {
	f.lastType = msgType
	f.lastAlert = alert
}

func (f *fakeBroadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request, initial []store.Alert) {
	w.WriteHeader(http.StatusNotImplemented)
}

func TestHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	NewServer(&fakeAlertStore{}, &fakeBroadcaster{}, "/ws/alerts").Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetAlertNotFound(t *testing.T) {
	t.Parallel()

	srv := NewServer(&fakeAlertStore{
		getAlert: func(ctx context.Context, alertID string) (store.Alert, error) {
			return store.Alert{}, store.ErrNotFound
		},
	}, &fakeBroadcaster{}, "/ws/alerts")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/missing", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestAcknowledgeBroadcasts(t *testing.T) {
	t.Parallel()

	want := store.Alert{
		AlertID: "550e8400-e29b-41d4-a716-446655440000",
		Status:  "Acknowledged",
	}
	bc := &fakeBroadcaster{}
	srv := NewServer(&fakeAlertStore{
		acknowledge: func(ctx context.Context, alertID, acknowledgedBy string) (store.Alert, error) {
			return want, nil
		},
	}, bc, "/ws/alerts")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/"+want.AlertID+"/acknowledge", strings.NewReader(`{"acknowledgedBy":"analyst@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if bc.lastType != "alert.acknowledged" || bc.lastAlert.AlertID != want.AlertID {
		t.Fatalf("broadcast = %q %+v", bc.lastType, bc.lastAlert)
	}
}

func TestListAlertsEmptyArray(t *testing.T) {
	t.Parallel()

	srv := NewServer(&fakeAlertStore{
		listAlerts: func(ctx context.Context, status, severity string, limit, offset int) ([]store.Alert, error) {
			return nil, nil
		},
	}, &fakeBroadcaster{}, "/ws/alerts")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var got []store.Alert
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty array, got %v", got)
	}
}

func TestListAlertsCapsLimit(t *testing.T) {
	t.Parallel()

	var gotLimit int
	srv := NewServer(&fakeAlertStore{
		listAlerts: func(ctx context.Context, status, severity string, limit, offset int) ([]store.Alert, error) {
			gotLimit = limit
			return []store.Alert{{AlertID: "a", DetectedAt: time.Now()}}, nil
		},
	}, &fakeBroadcaster{}, "/ws/alerts")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?limit=500", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if gotLimit != 200 {
		t.Fatalf("limit = %d", gotLimit)
	}
}
