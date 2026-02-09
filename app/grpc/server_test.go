package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
	"github.com/vibast-solutions/ms-go-profile/app/service"
	"github.com/vibast-solutions/ms-go-profile/app/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcRepoStub struct {
	createFn       func(ctx context.Context, profile *entity.Profile) error
	findByIDFn     func(ctx context.Context, id uint64) (*entity.Profile, error)
	findByUserIDFn func(ctx context.Context, userID uint64) (*entity.Profile, error)
	updateFn       func(ctx context.Context, profile *entity.Profile) error
	deleteFn       func(ctx context.Context, id uint64) error
}

func (s *grpcRepoStub) Create(ctx context.Context, profile *entity.Profile) error {
	if s.createFn != nil {
		return s.createFn(ctx, profile)
	}
	return nil
}

func (s *grpcRepoStub) FindByID(ctx context.Context, id uint64) (*entity.Profile, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *grpcRepoStub) FindByUserID(ctx context.Context, userID uint64) (*entity.Profile, error) {
	if s.findByUserIDFn != nil {
		return s.findByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *grpcRepoStub) Update(ctx context.Context, profile *entity.Profile) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, profile)
	}
	return nil
}

func (s *grpcRepoStub) Delete(ctx context.Context, id uint64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func newGRPCServerWithRepo(repo *grpcRepoStub) *ProfileServer {
	svc := service.NewProfileService(repo)
	return NewProfileServer(svc)
}

func TestCreateProfileInvalidArgument(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.CreateProfile(context.Background(), &types.CreateProfileRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %s", status.Code(err))
	}
}

func TestCreateProfileAlreadyExists(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByUserIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 9, UserID: 77}, nil
		},
	})

	_, err := server.CreateProfile(context.Background(), &types.CreateProfileRequest{
		UserId: 77,
		Email:  "john@example.com",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected codes.AlreadyExists, got %s", status.Code(err))
	}
}

func TestGetProfileNotFound(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.GetProfile(context.Background(), &types.GetProfileRequest{Id: 100})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected codes.NotFound, got %s", status.Code(err))
	}
}

func TestUpdateProfileSuccess(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return &entity.Profile{
				ID:     11,
				UserID: 44,
				Email:  "old@example.com",
			}, nil
		},
	})

	resp, err := server.UpdateProfile(context.Background(), &types.UpdateProfileRequest{
		Id:    11,
		Email: "new@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.GetEmail() != "new@example.com" {
		t.Fatalf("expected updated email, got %q", resp.GetEmail())
	}
}

func TestDeleteProfileNotFound(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return repository.ErrProfileNotFound
		},
	})

	_, err := server.DeleteProfile(context.Background(), &types.DeleteProfileRequest{Id: 15})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected codes.NotFound, got %s", status.Code(err))
	}
}

func TestDeleteProfileInternal(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		deleteFn: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	})

	_, err := server.DeleteProfile(context.Background(), &types.DeleteProfileRequest{Id: 15})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestCreateProfileSuccess(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		createFn: func(_ context.Context, profile *entity.Profile) error {
			profile.ID = 123
			return nil
		},
	})

	resp, err := server.CreateProfile(context.Background(), &types.CreateProfileRequest{
		UserId: 7,
		Email:  "john@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.GetId() != 123 || resp.GetUserId() != 7 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateProfileInternal(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		createFn: func(_ context.Context, _ *entity.Profile) error {
			return errors.New("db down")
		},
	})

	_, err := server.CreateProfile(context.Background(), &types.CreateProfileRequest{
		UserId: 7,
		Email:  "john@example.com",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestGetProfileInvalidArgument(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.GetProfile(context.Background(), &types.GetProfileRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %s", status.Code(err))
	}
}

func TestGetProfileInternal(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return nil, errors.New("db down")
		},
	})

	_, err := server.GetProfile(context.Background(), &types.GetProfileRequest{Id: 1})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestGetProfileByUserIDInvalidArgument(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.GetProfileByUserID(context.Background(), &types.GetProfileByUserIDRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %s", status.Code(err))
	}
}

func TestGetProfileByUserIDNotFound(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.GetProfileByUserID(context.Background(), &types.GetProfileByUserIDRequest{UserId: 1})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected codes.NotFound, got %s", status.Code(err))
	}
}

func TestGetProfileByUserIDSuccess(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByUserIDFn: func(_ context.Context, userID uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: 2, UserID: userID, Email: "john@example.com"}, nil
		},
	})

	resp, err := server.GetProfileByUserID(context.Background(), &types.GetProfileByUserIDRequest{UserId: 42})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.GetUserId() != 42 || resp.GetId() != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetProfileByUserIDInternal(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByUserIDFn: func(_ context.Context, _ uint64) (*entity.Profile, error) {
			return nil, errors.New("db down")
		},
	})

	_, err := server.GetProfileByUserID(context.Background(), &types.GetProfileByUserIDRequest{UserId: 42})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestUpdateProfileInvalidArgument(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.UpdateProfile(context.Background(), &types.UpdateProfileRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %s", status.Code(err))
	}
}

func TestUpdateProfileNotFound(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.UpdateProfile(context.Background(), &types.UpdateProfileRequest{
		Id:    10,
		Email: "x@example.com",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected codes.NotFound, got %s", status.Code(err))
	}
}

func TestUpdateProfileInternal(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{
		findByIDFn: func(_ context.Context, id uint64) (*entity.Profile, error) {
			return &entity.Profile{ID: id, UserID: 1, Email: "old@example.com"}, nil
		},
		updateFn: func(_ context.Context, _ *entity.Profile) error {
			return errors.New("write failed")
		},
	})
	_, err := server.UpdateProfile(context.Background(), &types.UpdateProfileRequest{
		Id:    10,
		Email: "x@example.com",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestDeleteProfileInvalidArgument(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	_, err := server.DeleteProfile(context.Background(), &types.DeleteProfileRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %s", status.Code(err))
	}
}

func TestDeleteProfileSuccess(t *testing.T) {
	server := newGRPCServerWithRepo(&grpcRepoStub{})
	resp, err := server.DeleteProfile(context.Background(), &types.DeleteProfileRequest{Id: 15})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.GetMessage() == "" {
		t.Fatal("expected non-empty delete message")
	}
}
