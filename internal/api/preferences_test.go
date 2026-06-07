package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/savedfilter"
	"github.com/google/uuid"
)

type fakePreferenceService struct {
	setUser   uuid.UUID
	setID     *uuid.UUID
	setCalled bool
}

func (f *fakePreferenceService) SetDefaultFilter(_ context.Context, userID uuid.UUID, filterID *uuid.UUID) error {
	f.setCalled = true
	f.setUser = userID
	f.setID = filterID
	return nil
}

func TestSetDefaultFilterRequiresAuth(t *testing.T) {
	h := &preferenceHandlers{prefs: &fakePreferenceService{}, filters: &fakeSavedFilterService{}}
	rec := httptest.NewRecorder()
	h.patch(rec, httptest.NewRequest(http.MethodPatch, "/api/preferences/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestSetDefaultFilterNotOwned(t *testing.T) {
	prefs := &fakePreferenceService{}
	h := &preferenceHandlers{prefs: prefs, filters: &fakeSavedFilterService{getByIDErr: savedfilter.ErrNotFound}}
	rec := httptest.NewRecorder()
	body := []byte(`{"defaultFilterId":"` + uuid.New().String() + `"}`)
	h.patch(rec, authedRequest(http.MethodPatch, "/api/preferences/", body, uuid.New()))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
	if prefs.setCalled {
		t.Fatal("SetDefaultFilter should not be called when the filter is not owned")
	}
}

func TestSetDefaultFilterValid(t *testing.T) {
	prefs := &fakePreferenceService{}
	h := &preferenceHandlers{prefs: prefs, filters: &fakeSavedFilterService{}}
	uid := uuid.New()
	id := uuid.New()
	rec := httptest.NewRecorder()
	h.patch(rec, authedRequest(http.MethodPatch, "/api/preferences/", []byte(`{"defaultFilterId":"`+id.String()+`"}`), uid))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !prefs.setCalled || prefs.setUser != uid || prefs.setID == nil || *prefs.setID != id {
		t.Fatalf("unexpected set: called=%v user=%s id=%v", prefs.setCalled, prefs.setUser, prefs.setID)
	}
}

func TestClearDefaultFilter(t *testing.T) {
	prefs := &fakePreferenceService{}
	h := &preferenceHandlers{prefs: prefs, filters: &fakeSavedFilterService{}}
	rec := httptest.NewRecorder()
	h.patch(rec, authedRequest(http.MethodPatch, "/api/preferences/", []byte(`{"defaultFilterId":null}`), uuid.New()))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !prefs.setCalled || prefs.setID != nil {
		t.Fatalf("clear should call set with nil id: called=%v id=%v", prefs.setCalled, prefs.setID)
	}
}
