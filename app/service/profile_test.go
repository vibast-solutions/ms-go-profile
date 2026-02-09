package service

import (
	"context"
	"errors"
	"testing"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

type mockCreateReq struct {
	userID uint64
	email  string
}

func (r mockCreateReq) GetUserId() uint64 { return r.userID }
func (r mockCreateReq) GetEmail() string  { return r.email }

type mockUpdateReq struct {
	id    uint64
	email string
}

func (r mockUpdateReq) GetId() uint64    { return r.id }
func (r mockUpdateReq) GetEmail() string { return r.email }

type mockRepo struct {
	createFn       func(ctx context.Context, profile *entity.Profile) error
	findByIDFn     func(ctx context.Context, id uint64) (*entity.Profile, error)
	findByUserIDFn func(ctx context.Context, userID uint64) (*entity.Profile, error)
	updateFn       func(ctx context.Context, profile *entity.Profile) error
	deleteFn       func(ctx context.Context, id uint64) error
}

func (m *mockRepo) Create(ctx context.Context, profile *entity.Profile) error {
	if m.createFn != nil {
		return m.createFn(ctx, profile)
	}
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id uint64) (*entity.Profile, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepo) FindByUserID(ctx context.Context, userID uint64) (*entity.Profile, error) {
	if m.findByUserIDFn != nil {
		return m.findByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepo) Update(ctx context.Context, profile *entity.Profile) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, profile)
	}
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestCreateSuccess(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, profile *entity.Profile) error {
			profile.ID = 99
			return nil
		},
	}
	svc := NewProfileService(repo)

	profile, err := svc.Create(context.Background(), mockCreateReq{
		userID: 42,
		email:  "john@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if profile.ID != 99 || profile.UserID != 42 || profile.Email != "john@example.com" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestCreateAlreadyExistsFromLookup(t *testing.T) {
	repo := &mockRepo{
		findByUserIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 1, UserID: 42}, nil
		},
	}
	svc := NewProfileService(repo)

	_, err := svc.Create(context.Background(), mockCreateReq{userID: 42, email: "john@example.com"})
	if !errors.Is(err, ErrProfileAlreadyExists) {
		t.Fatalf("expected ErrProfileAlreadyExists, got: %v", err)
	}
}

func TestCreateAlreadyExistsFromRepositoryError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *entity.Profile) error {
			return repository.ErrProfileAlreadyExists
		},
	}
	svc := NewProfileService(repo)

	_, err := svc.Create(context.Background(), mockCreateReq{userID: 42, email: "john@example.com"})
	if !errors.Is(err, ErrProfileAlreadyExists) {
		t.Fatalf("expected ErrProfileAlreadyExists, got: %v", err)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	svc := NewProfileService(&mockRepo{})
	_, err := svc.GetByID(context.Background(), 1)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestGetByUserIDNotFound(t *testing.T) {
	svc := NewProfileService(&mockRepo{})
	_, err := svc.GetByUserID(context.Background(), 1)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestUpdateNotFound(t *testing.T) {
	svc := NewProfileService(&mockRepo{})
	_, err := svc.Update(context.Background(), mockUpdateReq{id: 22, email: "new@example.com"})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestUpdateRepositoryNotFoundMapped(t *testing.T) {
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 22, UserID: 7, Email: "old@example.com"}, nil
		},
		updateFn: func(_ context.Context, _ *entity.Profile) error {
			return repository.ErrProfileNotFound
		},
	}
	svc := NewProfileService(repo)

	_, err := svc.Update(context.Background(), mockUpdateReq{id: 22, email: "new@example.com"})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}

func TestDeleteRepositoryNotFoundMapped(t *testing.T) {
	repo := &mockRepo{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrProfileNotFound
		},
	}
	svc := NewProfileService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got: %v", err)
	}
}
