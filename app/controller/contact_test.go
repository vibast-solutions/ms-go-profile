package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
	"github.com/vibast-solutions/ms-go-profile/app/service"
)

type contactRepoStub struct {
	createFn   func(ctx context.Context, contact *entity.Contact) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Contact, error)
	updateFn   func(ctx context.Context, contact *entity.Contact) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Contact, uint64, error)
}

func (s *contactRepoStub) Create(ctx context.Context, contact *entity.Contact) error {
	if s.createFn != nil {
		return s.createFn(ctx, contact)
	}
	return nil
}

func (s *contactRepoStub) FindByID(ctx context.Context, id uint64) (*entity.Contact, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *contactRepoStub) Update(ctx context.Context, contact *entity.Contact) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, contact)
	}
	return nil
}

func (s *contactRepoStub) Delete(ctx context.Context, id uint64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *contactRepoStub) List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Contact, uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, profileID, limit, offset)
	}
	return nil, 0, nil
}

func newContactControllerWithRepo(repo *contactRepoStub) *ContactController {
	svc := service.NewContactService(repo)
	return NewContactController(svc)
}

func TestContactCreateBadRequest(t *testing.T) {
	ctrl := newContactControllerWithRepo(&contactRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBufferString("{invalid"))
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

func TestContactCreateSuccess(t *testing.T) {
	ctrl := newContactControllerWithRepo(&contactRepoStub{
		createFn: func(_ context.Context, contact *entity.Contact) error {
			contact.ID = 15
			return nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBufferString(`{"profile_id":4}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.Create(ctx); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestContactGetByIDNotFound(t *testing.T) {
	ctrl := newContactControllerWithRepo(&contactRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/contacts/9", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("9")

	if err := ctrl.GetByID(ctx); err != nil {
		t.Fatalf("GetByID() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestContactDeleteNotFound(t *testing.T) {
	ctrl := newContactControllerWithRepo(&contactRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrContactNotFound
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/contacts/8", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("8")

	if err := ctrl.Delete(ctx); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestContactListInvalidQuery(t *testing.T) {
	ctrl := newContactControllerWithRepo(&contactRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/contacts?page=bad", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.List(ctx); err != nil {
		t.Fatalf("List() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestContactListSuccess(t *testing.T) {
	now := time.Now()
	dob := time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)
	ctrl := newContactControllerWithRepo(&contactRepoStub{
		listFn: func(_ context.Context, profileID uint64, limit, offset uint32) ([]*entity.Contact, uint64, error) {
			if profileID != 4 || limit != 5 || offset != 5 {
				t.Fatalf("unexpected list args profileID=%d limit=%d offset=%d", profileID, limit, offset)
			}
			return []*entity.Contact{
				{
					ID:        1,
					FirstName: "John",
					LastName:  "Doe",
					NIN:       "1234",
					DOB:       &dob,
					Phone:     "123456",
					CreatedAt: now,
					UpdatedAt: now,
					ProfileID: 4,
				},
			}, 1, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/contacts?profile_id=4&page=2&page_size=5", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.List(ctx); err != nil {
		t.Fatalf("List() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if payload["total"] == nil {
		t.Fatalf("expected paginated payload, got: %s", rec.Body.String())
	}
}
