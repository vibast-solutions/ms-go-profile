//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	authpb "github.com/vibast-solutions/ms-go-auth/app/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultProfileCallerAPIKey   = "profile-caller-key"
	defaultProfileNoAccessAPIKey = "profile-no-access-key"
	defaultProfileAppAPIKey      = "profile-app-api-key"
	profileAuthMockAddr          = "0.0.0.0:38081"
)

func profileCallerAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("PROFILE_CALLER_API_KEY")); value != "" {
		return value
	}
	return defaultProfileCallerAPIKey
}

func profileNoAccessAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("PROFILE_NO_ACCESS_API_KEY")); value != "" {
		return value
	}
	return defaultProfileNoAccessAPIKey
}

func profileAppAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("PROFILE_APP_API_KEY")); value != "" {
		return value
	}
	return defaultProfileAppAPIKey
}

type profileAuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
}

func (s *profileAuthGRPCServer) ValidateInternalAccess(ctx context.Context, req *authpb.ValidateInternalAccessRequest) (*authpb.ValidateInternalAccessResponse, error) {
	if incomingAPIKey(ctx) != profileAppAPIKey() {
		return nil, status.Error(codes.Unauthenticated, "unauthorized caller")
	}

	apiKey := strings.TrimSpace(req.GetApiKey())
	switch apiKey {
	case profileCallerAPIKey():
		return &authpb.ValidateInternalAccessResponse{
			ServiceName:   "profile-gateway",
			AllowedAccess: []string{"profile-service", "notifications-service"},
		}, nil
	case profileNoAccessAPIKey():
		return &authpb.ValidateInternalAccessResponse{
			ServiceName:   "profile-gateway",
			AllowedAccess: []string{"notifications-service"},
		}, nil
	default:
		return nil, status.Error(codes.Unauthenticated, "invalid api key")
	}
}

func incomingAPIKey(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("x-api-key")
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func TestMain(m *testing.M) {
	if os.Getenv("PROFILE_CALLER_API_KEY") == "" {
		_ = os.Setenv("PROFILE_CALLER_API_KEY", defaultProfileCallerAPIKey)
	}
	if os.Getenv("PROFILE_NO_ACCESS_API_KEY") == "" {
		_ = os.Setenv("PROFILE_NO_ACCESS_API_KEY", defaultProfileNoAccessAPIKey)
	}
	if os.Getenv("PROFILE_APP_API_KEY") == "" {
		_ = os.Setenv("PROFILE_APP_API_KEY", defaultProfileAppAPIKey)
	}

	listener, err := net.Listen("tcp", profileAuthMockAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start profile auth grpc mock: %v\n", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, &profileAuthGRPCServer{})

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	exitCode := m.Run()

	grpcServer.GracefulStop()
	_ = listener.Close()

	os.Exit(exitCode)
}
