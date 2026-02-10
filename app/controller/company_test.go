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

type companyRepoStub struct {
	createFn   func(ctx context.Context, company *entity.Company) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Company, error)
	updateFn   func(ctx context.Context, company *entity.Company) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error)
}

func (s *companyRepoStub) Create(ctx context.Context, company *entity.Company) error {
	if s.createFn != nil {
		return s.createFn(ctx, company)
	}
	return nil
}

func (s *companyRepoStub) FindByID(ctx context.Context, id uint64) (*entity.Company, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *companyRepoStub) Update(ctx context.Context, company *entity.Company) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, company)
	}
	return nil
}

func (s *companyRepoStub) Delete(ctx context.Context, id uint64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *companyRepoStub) List(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, profileID, companyType, limit, offset)
	}
	return nil, 0, nil
}

func newCompanyControllerWithRepo(repo *companyRepoStub) *CompanyController {
	svc := service.NewCompanyService(repo)
	return NewCompanyController(svc)
}

func TestCompanyCreateBadRequest(t *testing.T) {
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBufferString("{invalid"))
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

func TestCompanyCreateSuccess(t *testing.T) {
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{
		createFn: func(_ context.Context, company *entity.Company) error {
			company.ID = 15
			return nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBufferString(`{"name":"ACME","registration_no":"REG-1","fiscal_code":"FISC-1","profile_id":4}`))
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

func TestCompanyGetByIDNotFound(t *testing.T) {
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/companies/9", nil)
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

func TestCompanyDeleteNotFound(t *testing.T) {
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrCompanyNotFound
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/companies/8", nil)
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

func TestCompanyListInvalidQuery(t *testing.T) {
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/companies?profile_id=bad", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ctrl.List(ctx); err != nil {
		t.Fatalf("List() returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCompanyListSuccess(t *testing.T) {
	now := time.Now()
	ctrl := newCompanyControllerWithRepo(&companyRepoStub{
		listFn: func(_ context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error) {
			if profileID != 7 || companyType != "vendor" || limit != 5 || offset != 5 {
				t.Fatalf("unexpected list args profileID=%d companyType=%q limit=%d offset=%d", profileID, companyType, limit, offset)
			}
			return []*entity.Company{
				{
					ID:             1,
					Name:           "ACME",
					RegistrationNo: "REG-1",
					FiscalCode:     "FISC-1",
					ProfileID:      7,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			}, 1, nil
		},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/companies?profile_id=7&page=2&page_size=5&type=vendor", nil)
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
