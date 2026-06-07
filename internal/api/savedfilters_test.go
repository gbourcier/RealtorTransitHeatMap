package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/savedfilter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeSavedFilterService struct {
	listUser    uuid.UUID
	listRows    []savedfilter.SavedFilter
	created     *savedfilter.SavedFilter
	createErr   error
	updateUser  uuid.UUID
	updateID    uuid.UUID
	updateErr   error
	deleteUser  uuid.UUID
	deleteID    uuid.UUID
	getByIDUser uuid.UUID
	getByIDID   uuid.UUID
	getByIDErr  error
}

func (f *fakeSavedFilterService) List(_ context.Context, userID uuid.UUID) ([]savedfilter.SavedFilter, error) {
	f.listUser = userID
	return f.listRows, nil
}

func (f *fakeSavedFilterService) Create(_ context.Context, sf *savedfilter.SavedFilter) error {
	if f.createErr != nil {
		return f.createErr
	}
	sf.ID = uuid.New()
	f.created = sf
	return nil
}

func (f *fakeSavedFilterService) Update(_ context.Context, userID, id uuid.UUID, _ savedfilter.SavedFilter) error {
	f.updateUser = userID
	f.updateID = id
	return f.updateErr
}

func (f *fakeSavedFilterService) Delete(_ context.Context, userID, id uuid.UUID) error {
	f.deleteUser = userID
	f.deleteID = id
	return nil
}

func (f *fakeSavedFilterService) GetByID(_ context.Context, userID, id uuid.UUID) (*savedfilter.SavedFilter, error) {
	f.getByIDUser = userID
	f.getByIDID = id
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	return &savedfilter.SavedFilter{ID: id, UserID: userID, Name: "preset"}, nil
}

func withURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateSavedFilterRequiresAuth(t *testing.T) {
	h := &savedFilterHandlers{svc: &fakeSavedFilterService{}}
	rec := httptest.NewRecorder()
	h.create(rec, httptest.NewRequest(http.MethodPost, "/api/saved-filters/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestCreateSavedFilterValidation(t *testing.T) {
	h := &savedFilterHandlers{svc: &fakeSavedFilterService{}}
	rec := httptest.NewRecorder()
	h.create(rec, authedRequest(http.MethodPost, "/api/saved-filters/", []byte(`{"name":"  "}`), uuid.New()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateSavedFilterNameTaken(t *testing.T) {
	h := &savedFilterHandlers{svc: &fakeSavedFilterService{createErr: savedfilter.ErrNameTaken}}
	rec := httptest.NewRecorder()
	h.create(rec, authedRequest(http.MethodPost, "/api/saved-filters/", []byte(`{"name":"Downtown"}`), uuid.New()))
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", rec.Code)
	}
}

func TestCreateSavedFilterUsesContextUser(t *testing.T) {
	svc := &fakeSavedFilterService{}
	h := &savedFilterHandlers{svc: svc}
	uid := uuid.New()
	rec := httptest.NewRecorder()
	h.create(rec, authedRequest(http.MethodPost, "/api/saved-filters/", []byte(`{"name":"Downtown","maxPrice":500000,"favoritesOnly":true}`), uid))
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}
	if svc.created == nil || svc.created.UserID != uid {
		t.Fatalf("create used wrong user: %+v", svc.created)
	}
}

func TestUpdateSavedFilterNotFound(t *testing.T) {
	h := &savedFilterHandlers{svc: &fakeSavedFilterService{updateErr: savedfilter.ErrNotFound}}
	rec := httptest.NewRecorder()
	r := withURLParam(authedRequest(http.MethodPatch, "/api/saved-filters/x", []byte(`{"name":"X"}`), uuid.New()), "id", uuid.New().String())
	h.update(rec, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestUpdateSavedFilterNameTaken(t *testing.T) {
	h := &savedFilterHandlers{svc: &fakeSavedFilterService{updateErr: savedfilter.ErrNameTaken}}
	rec := httptest.NewRecorder()
	r := withURLParam(authedRequest(http.MethodPatch, "/api/saved-filters/x", []byte(`{"name":"X"}`), uuid.New()), "id", uuid.New().String())
	h.update(rec, r)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", rec.Code)
	}
}

func TestUpdateSavedFilterScope(t *testing.T) {
	svc := &fakeSavedFilterService{}
	h := &savedFilterHandlers{svc: svc}
	uid := uuid.New()
	id := uuid.New()
	rec := httptest.NewRecorder()
	r := withURLParam(authedRequest(http.MethodPatch, "/api/saved-filters/x", []byte(`{"name":"X"}`), uid), "id", id.String())
	h.update(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.updateUser != uid || svc.updateID != id {
		t.Fatalf("update used wrong scope: user=%s id=%s", svc.updateUser, svc.updateID)
	}
}

func TestDeleteSavedFilterNoContent(t *testing.T) {
	svc := &fakeSavedFilterService{}
	h := &savedFilterHandlers{svc: svc}
	uid := uuid.New()
	id := uuid.New()
	rec := httptest.NewRecorder()
	r := withURLParam(authedRequest(http.MethodDelete, "/api/saved-filters/x", nil, uid), "id", id.String())
	h.delete(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if svc.deleteUser != uid || svc.deleteID != id {
		t.Fatalf("delete used wrong scope: user=%s id=%s", svc.deleteUser, svc.deleteID)
	}
}

func TestListSavedFiltersScope(t *testing.T) {
	svc := &fakeSavedFilterService{}
	h := &savedFilterHandlers{svc: svc}
	uid := uuid.New()
	rec := httptest.NewRecorder()
	h.list(rec, authedRequest(http.MethodGet, "/api/saved-filters/", nil, uid))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.listUser != uid {
		t.Fatalf("list used wrong user: want %s got %s", uid, svc.listUser)
	}
}
