package service

import (
	"context"
	"errors"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
)

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileAlreadyExists = errors.New("profile already exists for this user")
)

type createProfileRequest interface {
	GetUserId() uint64
	GetEmail() string
}

type updateProfileRequest interface {
	GetId() uint64
	GetEmail() string
}

type ProfileService struct {
	profileRepo profileRepository
}

type profileRepository interface {
	Create(ctx context.Context, profile *entity.Profile) error
	FindByID(ctx context.Context, id uint64) (*entity.Profile, error)
	FindByUserID(ctx context.Context, userID uint64) (*entity.Profile, error)
	Update(ctx context.Context, profile *entity.Profile) error
	Delete(ctx context.Context, id uint64) error
}

func NewProfileService(profileRepo profileRepository) *ProfileService {
	return &ProfileService{profileRepo: profileRepo}
}

func (s *ProfileService) Create(ctx context.Context, req createProfileRequest) (*entity.Profile, error) {
	existing, err := s.profileRepo.FindByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrProfileAlreadyExists
	}

	now := time.Now()
	profile := &entity.Profile{
		UserID:    req.GetUserId(),
		Email:     req.GetEmail(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		if errors.Is(err, repository.ErrProfileAlreadyExists) {
			return nil, ErrProfileAlreadyExists
		}
		return nil, err
	}

	return profile, nil
}

func (s *ProfileService) GetByID(ctx context.Context, id uint64) (*entity.Profile, error) {
	profile, err := s.profileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}
	return profile, nil
}

func (s *ProfileService) GetByUserID(ctx context.Context, userId uint64) (*entity.Profile, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}
	return profile, nil
}

func (s *ProfileService) Update(ctx context.Context, req updateProfileRequest) (*entity.Profile, error) {
	profile, err := s.profileRepo.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	profile.Email = req.GetEmail()
	if err := s.profileRepo.Update(ctx, profile); err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	return profile, nil
}

func (s *ProfileService) Delete(ctx context.Context, id uint64) error {
	if err := s.profileRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	return nil
}
