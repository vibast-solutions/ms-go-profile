//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultHTTPBase = "http://localhost:28080"
	defaultGRPCAddr = "localhost:29090"
)

type httpClient struct {
	baseURL string
	client  *http.Client
}

func newHTTPClient(baseURL string) *httpClient {
	return &httpClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *httpClient) doJSON(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	return c.doJSONWithAPIKey(t, method, path, body, profileCallerAPIKey())
}

func (c *httpClient) doJSONWithAPIKey(t *testing.T, method, path string, body any, apiKey string) (*http.Response, []byte) {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}
		reqBody = bytes.NewReader(data)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioReadAll(resp)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	return resp, bodyBytes
}

func waitForHTTP(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodGet, baseURL+"/health", nil)
		req.Header.Set("X-API-Key", profileCallerAPIKey())
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("http service not ready at %s", baseURL)
}

func withGRPCAPIKey() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", profileCallerAPIKey())
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

func dialProfileGRPC(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), withGRPCAPIKey())
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}

	return conn
}

func dialProfileGRPCRaw(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}

	return conn
}

func grpcContextWithAPIKey(apiKey string) context.Context {
	if apiKey == "" {
		return context.Background()
	}
	return metadata.AppendToOutgoingContext(context.Background(), "x-api-key", apiKey)
}

func waitForGRPC(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("grpc service not ready at %s", addr)
}

func TestProfileE2E_CrossTransportCRUD(t *testing.T) {
	httpBase := os.Getenv("PROFILE_HTTP_URL")
	if httpBase == "" {
		httpBase = defaultHTTPBase
	}
	grpcAddr := os.Getenv("PROFILE_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = defaultGRPCAddr
	}

	if err := waitForHTTP(httpBase, 30*time.Second); err != nil {
		t.Fatalf("http not ready: %v", err)
	}
	if err := waitForGRPC(grpcAddr, 30*time.Second); err != nil {
		t.Fatalf("grpc not ready: %v", err)
	}

	httpClient := newHTTPClient(httpBase)

	conn := dialProfileGRPC(t, grpcAddr)
	defer conn.Close()
	grpcClient := types.NewProfileServiceClient(conn)

	rawConn := dialProfileGRPCRaw(t, grpcAddr)
	defer rawConn.Close()
	rawGRPCClient := types.NewProfileServiceClient(rawConn)

	state := struct {
		profile  *types.ProfileResponse
		userID   uint64
		emailV1  string
		emailV2  string
		deleteID uint64
	}{
		userID:  uint64(time.Now().UnixNano()%1_000_000) + 1000,
		emailV1: fmt.Sprintf("profile-e2e-%d@example.com", time.Now().UnixNano()),
	}
	state.emailV2 = "updated-" + state.emailV1

	t.Run("HTTPUnauthorizedMissingAPIKey", func(t *testing.T) {
		resp, _ := httpClient.doJSONWithAPIKey(t, http.MethodGet, "/health", nil, "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for missing x-api-key, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPForbiddenInsufficientAccess", func(t *testing.T) {
		resp, _ := httpClient.doJSONWithAPIKey(t, http.MethodGet, "/health", nil, profileNoAccessAPIKey())
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for insufficient access, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCUnauthorizedMissingAPIKey", func(t *testing.T) {
		_, err := rawGRPCClient.GetProfile(context.Background(), &types.GetProfileRequest{Id: 1})
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("expected Unauthenticated, got %v", err)
		}
	})

	t.Run("GRPCForbiddenInsufficientAccess", func(t *testing.T) {
		_, err := rawGRPCClient.GetProfile(grpcContextWithAPIKey(profileNoAccessAPIKey()), &types.GetProfileRequest{Id: 1})
		if status.Code(err) != codes.PermissionDenied {
			t.Fatalf("expected PermissionDenied, got %v", err)
		}
	})

	t.Run("HTTPValidationCreate", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/profiles", map[string]any{})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid create request, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCValidationCreate", func(t *testing.T) {
		_, err := grpcClient.CreateProfile(context.Background(), &types.CreateProfileRequest{})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("HTTPCreate", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/profiles", map[string]any{
			"user_id": state.userID,
			"email":   state.emailV1,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}

		var created types.ProfileResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create response failed: %v body=%s", err, string(body))
		}
		if created.GetId() == 0 || created.GetUserId() != state.userID || created.GetEmail() != state.emailV1 {
			t.Fatalf("unexpected create response: %+v", created)
		}

		state.profile = &created
		state.deleteID = created.GetId()
	})

	t.Run("HTTPCreateDuplicate", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/profiles", map[string]any{
			"user_id": state.userID,
			"email":   "duplicate-" + state.emailV1,
		})
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409 for duplicate user profile, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCGetByUserIDAfterHTTPCreate", func(t *testing.T) {
		got, err := grpcClient.GetProfileByUserID(context.Background(), &types.GetProfileByUserIDRequest{
			UserId: state.userID,
		})
		if err != nil {
			t.Fatalf("grpc get by user id failed: %v", err)
		}
		if got.GetId() != state.profile.GetId() || got.GetEmail() != state.emailV1 {
			t.Fatalf("unexpected grpc profile data: %+v", got)
		}
	})

	t.Run("GRPCUpdate", func(t *testing.T) {
		updated, err := grpcClient.UpdateProfile(context.Background(), &types.UpdateProfileRequest{
			Id:    state.profile.GetId(),
			Email: state.emailV2,
		})
		if err != nil {
			t.Fatalf("grpc update failed: %v", err)
		}
		if updated.GetEmail() != state.emailV2 {
			t.Fatalf("expected updated email %q, got %q", state.emailV2, updated.GetEmail())
		}
	})

	t.Run("HTTPGetByIDAfterGRPCUpdate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/profiles/"+strconv.FormatUint(state.profile.GetId(), 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var got types.ProfileResponse
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal get by id response failed: %v body=%s", err, string(body))
		}
		if got.GetEmail() != state.emailV2 {
			t.Fatalf("expected email %q, got %q", state.emailV2, got.GetEmail())
		}
	})

	t.Run("HTTPGetByUserID", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/profiles/user/"+strconv.FormatUint(state.userID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
	})

	t.Run("GRPCDelete", func(t *testing.T) {
		_, err := grpcClient.DeleteProfile(context.Background(), &types.DeleteProfileRequest{
			Id: state.deleteID,
		})
		if err != nil {
			t.Fatalf("grpc delete failed: %v", err)
		}
	})

	t.Run("HTTPDeleteAgainNotFound", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodDelete,
			"/profiles/"+strconv.FormatUint(state.deleteID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 after deleting already deleted profile, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPGetByIDNotFound", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodGet,
			"/profiles/"+strconv.FormatUint(state.deleteID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for deleted profile, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCGetNotFound", func(t *testing.T) {
		_, err := grpcClient.GetProfile(context.Background(), &types.GetProfileRequest{Id: state.deleteID})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected NotFound after delete, got %v", err)
		}
	})
}

func ioReadAll(resp *http.Response) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(resp.Body)
	return buf.Bytes(), err
}
