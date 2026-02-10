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

type addressRepoStub struct {
	createFn   func(ctx context.Context, address *entity.Address) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Address, error)
	updateFn   func(ctx context.Context, address *entity.Address) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Address, uint64, error)
}

func (s *addressRepoStub) Create(ctx context.Context, address *entity.Address) error {
	if s.createFn != nil {
		return s.createFn(ctx, address)
	}
	return nil
}

func (s *addressRepoStub) FindByID(ctx context.Context, id uint64) (*entity.Address, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *addressRepoStub) Update(ctx context.Context, address *entity.Address) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, address)
	}
	return nil
}

func (s *addressRepoStub) Delete(ctx context.Context, id uint64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *addressRepoStub) List(ctx context.Context, profileID uint64, limit, offset uint32) ([]*entity.Address, uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, profileID, limit, offset)
	}
	return nil, 0, nil
}

func newAddressControllerWithRepo(repo *addressRepoStub) *AddressController {
	svc := service.NewAddressService(repo)
	return NewAddressController(svc)
}

func TestAddressCreateBadRequest(t *testing.T) {
	ctrl := newAddressControllerWithRepo(&addressRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/addresses", bytes.NewBufferString("{invalid"))
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

func TestAddressCreateSuccess(t *testing.T) {
	ctrl := newAddressControllerWithRepo(&addressRepoStub{
		createFn: func(_ context.Context, address *entity.Address) error {
			address.ID = 5
			return nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/addresses", bytes.NewBufferString(`{"street_name":"Street","streen_no":"10","city":"City","county":"County","country":"Country","profile_id":7}`))
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

func TestAddressGetByIDNotFound(t *testing.T) {
	ctrl := newAddressControllerWithRepo(&addressRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/addresses/9", nil)
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

func TestAddressDeleteNotFound(t *testing.T) {
	ctrl := newAddressControllerWithRepo(&addressRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrAddressNotFound
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/addresses/8", nil)
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

func TestAddressListInvalidQuery(t *testing.T) {
	ctrl := newAddressControllerWithRepo(&addressRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/addresses?profile_id=bad", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.List(ctx); err != nil {
		t.Fatalf("List() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAddressListSuccess(t *testing.T) {
	now := time.Now()
	ctrl := newAddressControllerWithRepo(&addressRepoStub{
		listFn: func(_ context.Context, profileID uint64, limit, offset uint32) ([]*entity.Address, uint64, error) {
			if profileID != 7 || limit != 5 || offset != 5 {
				t.Fatalf("unexpected list args profileID=%d limit=%d offset=%d", profileID, limit, offset)
			}
			return []*entity.Address{
				{
					ID:         1,
					StreetName: "Street",
					StreenNo:   "10",
					City:       "City",
					County:     "County",
					Country:    "Country",
					ProfileID:  7,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			}, 1, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/addresses?profile_id=7&page=2&page_size=5", nil)
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
