package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

type mockCreateContactReq struct {
	firstName string
	lastName  string
	nin       string
	dob       string
	phone     string
	profileID uint64
	kind      string
}

func (r mockCreateContactReq) GetFirstName() string { return r.firstName }
func (r mockCreateContactReq) GetLastName() string  { return r.lastName }
func (r mockCreateContactReq) GetNin() string       { return r.nin }
func (r mockCreateContactReq) GetDob() string       { return r.dob }
func (r mockCreateContactReq) GetPhone() string     { return r.phone }
func (r mockCreateContactReq) GetProfileId() uint64 { return r.profileID }
func (r mockCreateContactReq) GetType() string      { return r.kind }

type mockUpdateContactReq struct {
	id        uint64
	firstName string
	lastName  string
	nin       string
	dob       string
	phone     string
	profileID uint64
	kind      string
}

func (r mockUpdateContactReq) GetId() uint64        { return r.id }
func (r mockUpdateContactReq) GetFirstName() string { return r.firstName }
func (r mockUpdateContactReq) GetLastName() string  { return r.lastName }
func (r mockUpdateContactReq) GetNin() string       { return r.nin }
func (r mockUpdateContactReq) GetDob() string       { return r.dob }
func (r mockUpdateContactReq) GetPhone() string     { return r.phone }
func (r mockUpdateContactReq) GetProfileId() uint64 { return r.profileID }
func (r mockUpdateContactReq) GetType() string      { return r.kind }

type mockListContactsReq struct {
	profileID uint64
	page      uint32
	pageSize  uint32
	kind      string
}

func (r mockListContactsReq) GetProfileId() uint64 { return r.profileID }
func (r mockListContactsReq) GetPage() uint32      { return r.page }
func (r mockListContactsReq) GetPageSize() uint32  { return r.pageSize }
func (r mockListContactsReq) GetType() string      { return r.kind }

type mockContactRepo struct {
	createFn   func(ctx context.Context, contact *entity.Contact) error
	findByIDFn func(ctx context.Context, id uint64) (*entity.Contact, error)
	updateFn   func(ctx context.Context, contact *entity.Contact) error
	deleteFn   func(ctx context.Context, id uint64) error
	listFn     func(ctx context.Context, profileID uint64, contactType string, limit, offset uint32) ([]*entity.Contact, uint64, error)
}

func (m *mockContactRepo) Create(ctx context.Context, contact *entity.Contact) error {
	if m.createFn != nil {
		return m.createFn(ctx, contact)
	}
	return nil
}

func (m *mockContactRepo) FindByID(ctx context.Context, id uint64) (*entity.Contact, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockContactRepo) Update(ctx context.Context, contact *entity.Contact) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, contact)
	}
	return nil
}

func (m *mockContactRepo) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockContactRepo) List(ctx context.Context, profileID uint64, contactType string, limit, offset uint32) ([]*entity.Contact, uint64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, profileID, contactType, limit, offset)
	}
	return nil, 0, nil
}

func TestContactCreateSuccess(t *testing.T) {
	repo := &mockContactRepo{
		createFn: func(_ context.Context, contact *entity.Contact) error {
			contact.ID = 17
			return nil
		},
	}
	svc := NewContactService(repo)

	contact, err := svc.Create(context.Background(), mockCreateContactReq{
		firstName: "John",
		lastName:  "Doe",
		nin:       "1234",
		dob:       "1990-01-02",
		phone:     "123456",
		profileID: 9,
		kind:      "friend",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if contact.ID != 17 || contact.FirstName != "John" || contact.ProfileID != 9 {
		t.Fatalf("unexpected contact: %+v", contact)
	}
}

func TestContactCreateInvalidDOB(t *testing.T) {
	svc := NewContactService(&mockContactRepo{})
	_, err := svc.Create(context.Background(), mockCreateContactReq{dob: "1990/01/02"})
	if err == nil {
		t.Fatal("expected parse error for invalid dob")
	}
}

func TestContactGetByIDNotFound(t *testing.T) {
	svc := NewContactService(&mockContactRepo{})
	_, err := svc.GetByID(context.Background(), 3)
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("expected ErrContactNotFound, got %v", err)
	}
}

func TestContactUpdateNotFound(t *testing.T) {
	svc := NewContactService(&mockContactRepo{})
	_, err := svc.Update(context.Background(), mockUpdateContactReq{
		id:        4,
		firstName: "John",
		lastName:  "Doe",
		nin:       "1234",
		dob:       "1990-01-02",
		phone:     "123456",
		profileID: 8,
	})
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("expected ErrContactNotFound, got %v", err)
	}
}

func TestContactDeleteRepositoryNotFoundMapped(t *testing.T) {
	repo := &mockContactRepo{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrContactNotFound
		},
	}
	svc := NewContactService(repo)

	err := svc.Delete(context.Background(), 10)
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("expected ErrContactNotFound, got %v", err)
	}
}

func TestContactListDefaults(t *testing.T) {
	now := time.Now()
	repo := &mockContactRepo{
		listFn: func(_ context.Context, profileID uint64, contactType string, limit, offset uint32) ([]*entity.Contact, uint64, error) {
			if profileID != 5 || contactType != "emergency" || limit != 20 || offset != 0 {
				t.Fatalf("unexpected list args profileID=%d contactType=%q limit=%d offset=%d", profileID, contactType, limit, offset)
			}
			return []*entity.Contact{{ID: 1, FirstName: "John", CreatedAt: now, UpdatedAt: now}}, 1, nil
		},
	}
	svc := NewContactService(repo)

	result, err := svc.List(context.Background(), mockListContactsReq{profileID: 5, page: 0, pageSize: 0, kind: "emergency"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 || result.PageSize != 20 || result.Total != 1 {
		t.Fatalf("unexpected list result: %+v", result)
	}
}
