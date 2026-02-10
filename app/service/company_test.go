package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

type mockCreateCompanyReq struct {
	name           string
	registrationNo string
	fiscalCode     string
	profileID      uint64
	kind           string
}

func (r mockCreateCompanyReq) GetName() string           { return r.name }
func (r mockCreateCompanyReq) GetRegistrationNo() string { return r.registrationNo }
func (r mockCreateCompanyReq) GetFiscalCode() string     { return r.fiscalCode }
func (r mockCreateCompanyReq) GetProfileId() uint64      { return r.profileID }
func (r mockCreateCompanyReq) GetType() string           { return r.kind }

type mockUpdateCompanyReq struct {
	id uint64
	mockCreateCompanyReq
}

func (r mockUpdateCompanyReq) GetId() uint64 { return r.id }

type mockListCompaniesReq struct {
	profileID uint64
	page      uint32
	pageSize  uint32
	kind      string
}

func (r mockListCompaniesReq) GetProfileId() uint64 { return r.profileID }
func (r mockListCompaniesReq) GetPage() uint32      { return r.page }
func (r mockListCompaniesReq) GetPageSize() uint32  { return r.pageSize }
func (r mockListCompaniesReq) GetType() string      { return r.kind }

type mockCompanyRepo struct {
	createFn   func(ctx context.Context, company *entity.Company) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Company, error)
	updateFn   func(ctx context.Context, company *entity.Company) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error)
}

func (m *mockCompanyRepo) Create(ctx context.Context, company *entity.Company) error {
	if m.createFn != nil {
		return m.createFn(ctx, company)
	}
	return nil
}

func (m *mockCompanyRepo) FindByID(ctx context.Context, id uint64) (*entity.Company, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockCompanyRepo) Update(ctx context.Context, company *entity.Company) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, company)
	}
	return nil
}

func (m *mockCompanyRepo) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockCompanyRepo) List(ctx context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, profileID, companyType, limit, offset)
	}
	return nil, 0, nil
}

func TestCompanyCreateSuccess(t *testing.T) {
	repo := &mockCompanyRepo{
		createFn: func(_ context.Context, company *entity.Company) error {
			company.ID = 23
			return nil
		},
	}
	svc := NewCompanyService(repo)

	company, err := svc.Create(context.Background(), mockCreateCompanyReq{
		name:           "ACME",
		registrationNo: "REG-1",
		fiscalCode:     "FISC-1",
		profileID:      9,
		kind:           "vendor",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if company.ID != 23 || company.Name != "ACME" || company.ProfileID != 9 {
		t.Fatalf("unexpected company: %+v", company)
	}
}

func TestCompanyGetByIDNotFound(t *testing.T) {
	svc := NewCompanyService(&mockCompanyRepo{})
	_, err := svc.GetByID(context.Background(), 3)
	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestCompanyUpdateNotFound(t *testing.T) {
	svc := NewCompanyService(&mockCompanyRepo{})
	_, err := svc.Update(context.Background(), mockUpdateCompanyReq{
		id: 4,
		mockCreateCompanyReq: mockCreateCompanyReq{
			name:           "ACME",
			registrationNo: "REG-1",
			fiscalCode:     "FISC-1",
			profileID:      9,
		},
	})
	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestCompanyDeleteRepositoryNotFoundMapped(t *testing.T) {
	repo := &mockCompanyRepo{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrCompanyNotFound
		},
	}
	svc := NewCompanyService(repo)

	err := svc.Delete(context.Background(), 10)
	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestCompanyListDefaults(t *testing.T) {
	now := time.Now()
	repo := &mockCompanyRepo{
		listFn: func(_ context.Context, profileID uint64, companyType string, limit, offset uint32) ([]*entity.Company, uint64, error) {
			if profileID != 7 || companyType != "vendor" || limit != 20 || offset != 0 {
				t.Fatalf("unexpected list args profileID=%d companyType=%q limit=%d offset=%d", profileID, companyType, limit, offset)
			}
			return []*entity.Company{{ID: 1, Name: "ACME", CreatedAt: now, UpdatedAt: now}}, 1, nil
		},
	}
	svc := NewCompanyService(repo)

	result, err := svc.List(context.Background(), mockListCompaniesReq{profileID: 7, page: 0, pageSize: 0, kind: "vendor"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 || result.PageSize != 20 || result.Total != 1 {
		t.Fatalf("unexpected list result: %+v", result)
	}
}
