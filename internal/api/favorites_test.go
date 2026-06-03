package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/auth"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/favorite"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeFavoriteService struct {
	addCalls  []favorite.Key
	addUser   uuid.UUID
	addErr    error
	removed   []favorite.Key
	batch     []favorite.Key
	batchUser uuid.UUID
	listUser  uuid.UUID
	listSort  favorite.Sort
	listRows  []favorite.Row
	listTotal int64
}

func (f *fakeFavoriteService) Add(_ context.Context, userID uuid.UUID, board, mls int) error {
	f.addUser = userID
	f.addCalls = append(f.addCalls, favorite.Key{Board: board, MLS: mls})
	return f.addErr
}

func (f *fakeFavoriteService) Remove(_ context.Context, _ uuid.UUID, board, mls int) error {
	f.removed = append(f.removed, favorite.Key{Board: board, MLS: mls})
	return nil
}

func (f *fakeFavoriteService) RemoveBatch(_ context.Context, userID uuid.UUID, keys []favorite.Key) (int64, error) {
	f.batchUser = userID
	f.batch = keys
	return int64(len(keys)), nil
}

func (f *fakeFavoriteService) List(_ context.Context, userID uuid.UUID, _ favorite.Page, sort favorite.Sort) ([]favorite.Row, int64, error) {
	f.listUser = userID
	f.listSort = sort
	return f.listRows, f.listTotal, nil
}

func authedRequest(method, target string, body []byte, userID uuid.UUID) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	return auth.WithUser(r, &user.User{ID: userID})
}

func TestAddFavoriteRequiresAuth(t *testing.T) {
	h := &favoriteHandlers{svc: &fakeFavoriteService{}}
	rec := httptest.NewRecorder()
	h.add(rec, httptest.NewRequest(http.MethodPost, "/api/favorites/", bytes.NewReader([]byte(`{"board":1,"mls":2}`))))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestAddFavoriteValidation(t *testing.T) {
	h := &favoriteHandlers{svc: &fakeFavoriteService{}}
	rec := httptest.NewRecorder()
	h.add(rec, authedRequest(http.MethodPost, "/api/favorites/", []byte(`{"board":0,"mls":2}`), uuid.New()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestAddFavoriteListingNotFound(t *testing.T) {
	h := &favoriteHandlers{svc: &fakeFavoriteService{addErr: favorite.ErrListingNotFound}}
	rec := httptest.NewRecorder()
	h.add(rec, authedRequest(http.MethodPost, "/api/favorites/", []byte(`{"board":1,"mls":2}`), uuid.New()))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestAddFavoriteUsesContextUser(t *testing.T) {
	svc := &fakeFavoriteService{}
	h := &favoriteHandlers{svc: svc}
	uid := uuid.New()
	rec := httptest.NewRecorder()
	h.add(rec, authedRequest(http.MethodPost, "/api/favorites/", []byte(`{"board":7,"mls":9}`), uid))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.addUser != uid {
		t.Fatalf("handler used wrong user id: want %s got %s", uid, svc.addUser)
	}
	if len(svc.addCalls) != 1 || svc.addCalls[0] != (favorite.Key{Board: 7, MLS: 9}) {
		t.Fatalf("unexpected add calls: %+v", svc.addCalls)
	}
}

func TestRemoveBatchEmpty(t *testing.T) {
	h := &favoriteHandlers{svc: &fakeFavoriteService{}}
	rec := httptest.NewRecorder()
	h.removeBatch(rec, authedRequest(http.MethodDelete, "/api/favorites/", []byte(`{"items":[]}`), uuid.New()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestRemoveBatchDeletesOnlyCallerScope(t *testing.T) {
	svc := &fakeFavoriteService{}
	h := &favoriteHandlers{svc: svc}
	uid := uuid.New()
	rec := httptest.NewRecorder()
	h.removeBatch(rec, authedRequest(http.MethodDelete, "/api/favorites/", []byte(`{"items":[{"board":1,"mls":2},{"board":3,"mls":4}]}`), uid))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var out struct {
		Deleted int64 `json:"deleted"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Deleted != 2 {
		t.Fatalf("want 2 deleted, got %d", out.Deleted)
	}
	if svc.batchUser != uid {
		t.Fatalf("batch used wrong user: want %s got %s", uid, svc.batchUser)
	}
}

func TestListFavoritesInvalidSort(t *testing.T) {
	h := &favoriteHandlers{svc: &fakeFavoriteService{}}
	rec := httptest.NewRecorder()
	h.list(rec, authedRequest(http.MethodGet, "/api/favorites/?sortBy=bogus", nil, uuid.New()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestListFavoritesDefaultSortAndScope(t *testing.T) {
	svc := &fakeFavoriteService{listTotal: 0, listRows: nil}
	h := &favoriteHandlers{svc: svc}
	uid := uuid.New()
	rec := httptest.NewRecorder()
	h.list(rec, authedRequest(http.MethodGet, "/api/favorites/", nil, uid))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if svc.listUser != uid {
		t.Fatalf("list used wrong user: want %s got %s", uid, svc.listUser)
	}
	if svc.listSort.By != "favorited_date" || svc.listSort.Dir != "desc" {
		t.Fatalf("unexpected default sort: %+v", svc.listSort)
	}
}

func TestRemoveOneNoContent(t *testing.T) {
	svc := &fakeFavoriteService{}
	h := &favoriteHandlers{svc: svc}
	r := authedRequest(http.MethodDelete, "/api/favorites/1/2", nil, uuid.New())
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("board", "1")
	rctx.URLParams.Add("mls", "2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.removeOne(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if len(svc.removed) != 1 || svc.removed[0] != (favorite.Key{Board: 1, MLS: 2}) {
		t.Fatalf("unexpected removed: %+v", svc.removed)
	}
}
