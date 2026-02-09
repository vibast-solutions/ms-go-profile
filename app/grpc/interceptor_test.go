package grpc

import (
	"context"
	"errors"
	"strings"
	"testing"

	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRequestIDInterceptorUsesIncomingID(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(requestIDHeader, "req-incoming"))
	info := &grpcpkg.UnaryServerInfo{FullMethod: "/profile.ProfileService/GetProfile"}

	_, err := RequestIDInterceptor()(ctx, nil, info, func(ctx context.Context, _ interface{}) (interface{}, error) {
		if got := RequestIDFromContext(ctx); got != "req-incoming" {
			t.Fatalf("expected request id req-incoming, got %q", got)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRequestIDInterceptorGeneratesID(t *testing.T) {
	info := &grpcpkg.UnaryServerInfo{FullMethod: "/profile.ProfileService/GetProfile"}

	_, err := RequestIDInterceptor()(context.Background(), nil, info, func(ctx context.Context, _ interface{}) (interface{}, error) {
		got := RequestIDFromContext(ctx)
		if got == "" || !strings.HasPrefix(got, "grpc-") {
			t.Fatalf("expected generated grpc-* request id, got %q", got)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRecoveryInterceptorConvertsPanicToInternal(t *testing.T) {
	info := &grpcpkg.UnaryServerInfo{FullMethod: "/profile.ProfileService/GetProfile"}
	_, err := RecoveryInterceptor()(context.Background(), nil, info, func(context.Context, interface{}) (interface{}, error) {
		panic("boom")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if status.Code(err) != codes.Internal {
		t.Fatalf("expected codes.Internal, got %s", status.Code(err))
	}
}

func TestLoggingInterceptorPassThrough(t *testing.T) {
	info := &grpcpkg.UnaryServerInfo{FullMethod: "/profile.ProfileService/GetProfile"}

	resp, err := LoggingInterceptor()(context.Background(), "req", info, func(context.Context, interface{}) (interface{}, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected response 'ok', got %#v", resp)
	}

	expectedErr := status.Error(codes.NotFound, "missing")
	_, err = LoggingInterceptor()(context.Background(), "req", info, func(context.Context, interface{}) (interface{}, error) {
		return nil, expectedErr
	})
	if !errors.Is(err, expectedErr) && status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestLoggerWithContextIncludesRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDContextKey{}, "req-42")
	entry := loggerWithContext(ctx)
	if got, ok := entry.Data["request_id"]; !ok || got != "req-42" {
		t.Fatalf("expected request_id=req-42 in logger entry, got %#v", entry.Data)
	}
}
