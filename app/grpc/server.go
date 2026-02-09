package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/service"
	"github.com/vibast-solutions/ms-go-profile/app/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileServer struct {
	types.UnimplementedProfileServiceServer
	profileService *service.ProfileService
}

func NewProfileServer(profileService *service.ProfileService) *ProfileServer {
	return &ProfileServer{profileService: profileService}
}

func (s *ProfileServer) CreateProfile(ctx context.Context, pbReq *types.CreateProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.WithField("user_id", pbReq.UserId).Debug("Create profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("user_id", pbReq.GetUserId()).Info("Create profile request received (grpc)")
	profile, err := s.profileService.Create(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "profile already exists for this user")
		}
		l.WithError(err).WithField("user_id", pbReq.GetUserId()).Error("Create profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", profile.ID).WithField("user_id", profile.UserID).Info("Profile created (grpc)")

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) GetProfile(ctx context.Context, pbReq *types.GetProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Get profile request received (grpc)")
	profile, err := s.profileService.GetByID(ctx, pbReq.GetId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Get profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) GetProfileByUserID(ctx context.Context, pbReq *types.GetProfileByUserIDRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Get profile by user ID validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("user_id", pbReq.GetUserId()).Info("Get profile by user ID request received (grpc)")
	profile, err := s.profileService.GetByUserID(ctx, pbReq.GetUserId())
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("user_id", pbReq.GetUserId()).Error("Get profile by user ID failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) UpdateProfile(ctx context.Context, pbReq *types.UpdateProfileRequest) (*types.ProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Update profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Update profile request received (grpc)")
	profile, err := s.profileService.Update(ctx, pbReq)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Update profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Profile updated (grpc)")
	return &types.ProfileResponse{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ProfileServer) DeleteProfile(ctx context.Context, pbReq *types.DeleteProfileRequest) (*types.DeleteProfileResponse, error) {
	l := loggerWithContext(ctx)
	if err := pbReq.Validate(); err != nil {
		l.Debug("Delete profile validation failed (grpc)")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Delete profile request received (grpc)")
	if err := s.profileService.Delete(ctx, pbReq.GetId()); err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		l.WithError(err).WithField("profile_id", pbReq.GetId()).Error("Delete profile failed (grpc)")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	l.WithField("profile_id", pbReq.GetId()).Info("Profile deleted (grpc)")
	return &types.DeleteProfileResponse{
		Message: "profile deleted successfully",
	}, nil
}
