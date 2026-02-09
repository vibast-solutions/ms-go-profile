package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
	"github.com/vibast-solutions/ms-go-profile/app/service"
)

type controllerRepoStub struct {
	createFn       func(ctx context.Context, profile *entity.Profile) error
	findByIDFn     func(ctx context.Context, id uint64) (*entity.Profile, error)
	findByUserIDFn func(ctx context.Context, userID uint64) (*entity.Profile, error)
	updateFn       func(ctx context.Context, profile *entity.Profile) error
	deleteFn       func(ctx context.Context, id uint64) error
}

func (s *controllerRepoStub) Create(ctx context.Context, profile *entity.Profile) error {
	if s.createFn != nil {
		return s.createFn(ctx, profile)
	}
	return nil
}

func (s *controllerRepoStub) FindByID(ctx context.Context, id uint64) (*entity.Profile, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *controllerRepoStub) FindByUserID(ctx context.Context, userID uint64) (*entity.Profile, error) {
	if s.findByUserIDFn != nil {
		return s.findByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *controllerRepoStub) Update(ctx context.Context, profile *entity.Profile) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, profile)
	}
	return nil
}

func (s *controllerRepoStub) Delete(ctx context.Context, id uint64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func newControllerWithRepo(repo *controllerRepoStub) *ProfileController {
	svc := service.NewProfileService(repo)
	return NewProfileController(svc)
}

func TestCreateBadRequest(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.Create(ctx); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateConflict(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		findByUserIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 1, UserID: 7}, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString(`{"user_id":7,"email":"john@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.Create(ctx); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCreateSuccess(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		createFn: func(_ context.Context, profile *entity.Profile) error {
			profile.ID = 101
			return nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString(`{"user_id":7,"email":"john@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.Create(ctx); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if payload["id"] == nil || payload["user_id"] == nil {
		t.Fatalf("expected profile response payload, got: %s", rec.Body.String())
	}
}

func TestGetByIDInvalidID(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profiles/x", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("x")

	if err := ctrl.GetByID(ctx); err != nil {
		t.Fatalf("GetByID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profiles/99", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("99")

	if err := ctrl.GetByID(ctx); err != nil {
		t.Fatalf("GetByID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrProfileNotFound
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/profiles/77", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("77")

	if err := ctrl.Delete(ctx); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetByUserIDInvalidID(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profiles/user/x", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("user_id")
	ctx.SetParamValues("x")

	if err := ctrl.GetByUserID(ctx); err != nil {
		t.Fatalf("GetByUserID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetByUserIDSuccess(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		findByUserIDFn: func(_ context.Context, userID uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 11, UserID: userID, Email: "john@example.com"}, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profiles/user/42", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("user_id")
	ctx.SetParamValues("42")

	if err := ctrl.GetByUserID(ctx); err != nil {
		t.Fatalf("GetByUserID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetByUserIDInternal(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		findByUserIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return nil, errors.New("db unavailable")
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profiles/user/42", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("user_id")
	ctx.SetParamValues("42")

	if err := ctrl.GetByUserID(ctx); err != nil {
		t.Fatalf("GetByUserID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateBadRequest(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/profiles/x", bytes.NewBufferString(`{"email":"a@b.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("x")

	if err := ctrl.Update(ctx); err != nil {
		t.Fatalf("Update() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateNotFound(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/profiles/10", bytes.NewBufferString(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	if err := ctrl.Update(ctx); err != nil {
		t.Fatalf("Update() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateInternal(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		findByIDFn: func(_ context.Context, id uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: id, UserID: 7, Email: "old@example.com"}, nil
		},
		updateFn: func(_ context.Context, _ *entity.Profile) error {
			return errors.New("write failed")
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/profiles/10", bytes.NewBufferString(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	if err := ctrl.Update(ctx); err != nil {
		t.Fatalf("Update() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateSuccess(t *testing.T) {
	ctrl := newControllerWithRepo(&controllerRepoStub{
		findByIDFn: func(_ context.Context, id uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: id, UserID: 7, Email: "old@example.com"}, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/profiles/10", bytes.NewBufferString(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	if err := ctrl.Update(ctx); err != nil {
		t.Fatalf("Update() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
