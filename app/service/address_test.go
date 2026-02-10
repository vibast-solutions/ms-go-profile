package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

type mockCreateAddressReq struct {
	streetName string
	streenNo   string
	city       string
	county     string
	country    string
	profileID  uint64
	postalCode string
	building   string
	apartment  string
	additional string
	kind       string
}

func (r mockCreateAddressReq) GetStreetName() string     { return r.streetName }
func (r mockCreateAddressReq) GetStreenNo() string       { return r.streenNo }
func (r mockCreateAddressReq) GetCity() string           { return r.city }
func (r mockCreateAddressReq) GetCounty() string         { return r.county }
func (r mockCreateAddressReq) GetCountry() string        { return r.country }
func (r mockCreateAddressReq) GetProfileId() uint64      { return r.profileID }
func (r mockCreateAddressReq) GetPostalCode() string     { return r.postalCode }
func (r mockCreateAddressReq) GetBuilding() string       { return r.building }
func (r mockCreateAddressReq) GetApartment() string      { return r.apartment }
func (r mockCreateAddressReq) GetAdditionalData() string { return r.additional }
func (r mockCreateAddressReq) GetType() string           { return r.kind }

type mockUpdateAddressReq struct {
	id uint64
	mockCreateAddressReq
}

func (r mockUpdateAddressReq) GetId() uint64 { return r.id }

type mockListAddressesReq struct {
	profileID uint64
	page      uint32
	pageSize  uint32
	kind      string
}

func (r mockListAddressesReq) GetProfileId() uint64 { return r.profileID }
func (r mockListAddressesReq) GetPage() uint32      { return r.page }
func (r mockListAddressesReq) GetPageSize() uint32  { return r.pageSize }
func (r mockListAddressesReq) GetType() string      { return r.kind }

type mockAddressRepo struct {
	createFn   func(ctx context.Context, address *entity.Address) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Address, error)
	updateFn   func(ctx context.Context, address *entity.Address) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, addressType string, limit, offset uint32) ([]*entity.Address, uint64, error)
}

func (m *mockAddressRepo) Create(ctx context.Context, address *entity.Address) error {
	if m.createFn != nil {
		return m.createFn(ctx, address)
	}
	return nil
}

func (m *mockAddressRepo) FindByID(ctx context.Context, id uint64) (*entity.Address, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAddressRepo) Update(ctx context.Context, address *entity.Address) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, address)
	}
	return nil
}

func (m *mockAddressRepo) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockAddressRepo) List(ctx context.Context, profileID uint64, addressType string, limit, offset uint32) ([]*entity.Address, uint64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, profileID, addressType, limit, offset)
	}
	return nil, 0, nil
}

func TestAddressCreateSuccess(t *testing.T) {
	repo := &mockAddressRepo{
		createFn: func(_ context.Context, address *entity.Address) error {
			address.ID = 13
			return nil
		},
	}
	svc := NewAddressService(repo)

	address, err := svc.Create(context.Background(), mockCreateAddressReq{
		streetName: "Street",
		streenNo:   "10",
		city:       "City",
		county:     "County",
		country:    "Country",
		profileID:  7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if address.ID != 13 || address.StreetName != "Street" || address.ProfileID != 7 {
		t.Fatalf("unexpected address: %+v", address)
	}
}

func TestAddressGetByIDNotFound(t *testing.T) {
	svc := NewAddressService(&mockAddressRepo{})
	_, err := svc.GetByID(context.Background(), 3)
	if !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
}

func TestAddressUpdateNotFound(t *testing.T) {
	svc := NewAddressService(&mockAddressRepo{})
	_, err := svc.Update(context.Background(), mockUpdateAddressReq{
		id: 1,
		mockCreateAddressReq: mockCreateAddressReq{
			streetName: "Street",
			streenNo:   "10",
			city:       "City",
			county:     "County",
			country:    "Country",
			profileID:  1,
		},
	})
	if !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
}

func TestAddressDeleteNotFoundMapped(t *testing.T) {
	repo := &mockAddressRepo{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrAddressNotFound
		},
	}
	svc := NewAddressService(repo)

	err := svc.Delete(context.Background(), 3)
	if !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
}

func TestAddressListDefaults(t *testing.T) {
	now := time.Now()
	repo := &mockAddressRepo{
		listFn: func(_ context.Context, profileID uint64, addressType string, limit, offset uint32) ([]*entity.Address, uint64, error) {
			if profileID != 7 || addressType != "billing" || limit != 20 || offset != 0 {
				t.Fatalf("unexpected list args profileID=%d addressType=%q limit=%d offset=%d", profileID, addressType, limit, offset)
			}
			return []*entity.Address{{ID: 1, StreetName: "Street", CreatedAt: now, UpdatedAt: now}}, 1, nil
		},
	}
	svc := NewAddressService(repo)

	result, err := svc.List(context.Background(), mockListAddressesReq{profileID: 7, page: 0, pageSize: 0, kind: "billing"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 || result.PageSize != 20 || result.Total != 1 {
		t.Fatalf("unexpected list result: %+v", result)
	}
}
