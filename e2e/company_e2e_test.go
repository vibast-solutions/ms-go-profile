//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/vibast-solutions/ms-go-profile/app/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCompanyE2E_CrossTransportCRUDAndList(t *testing.T) {
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

	state := struct {
		profileAID uint64
		profileBID uint64

		companyHTTPID uint64
		companyGRPCID uint64
		companyBID    uint64
	}{}

	nowNanos := time.Now().UnixNano()
	userA := uint64(nowNanos%1_000_000) + 30_000
	userB := userA + 1
	emailA := fmt.Sprintf("company-e2e-a-%d@example.com", nowNanos)
	emailB := fmt.Sprintf("company-e2e-b-%d@example.com", nowNanos)

	createProfileHTTP := func(t *testing.T, userID uint64, email string) uint64 {
		t.Helper()
		resp, body := httpClient.doJSON(t, http.MethodPost, "/profiles", map[string]any{
			"user_id": userID,
			"email":   email,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 for profile create, got %d body=%s", resp.StatusCode, string(body))
		}
		var created types.ProfileResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal profile create failed: %v body=%s", err, string(body))
		}
		return created.GetId()
	}

	createProfileGRPC := func(t *testing.T, userID uint64, email string) uint64 {
		t.Helper()
		created, err := grpcClient.CreateProfile(context.Background(), &types.CreateProfileRequest{
			UserId: userID,
			Email:  email,
		})
		if err != nil {
			t.Fatalf("grpc create profile failed: %v", err)
		}
		return created.GetId()
	}

	t.Run("HTTPCreateValidationMissingMandatory", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/companies", map[string]any{
			"profile_id": 1,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCCreateValidationMissingMandatory", func(t *testing.T) {
		_, err := grpcClient.CreateCompany(context.Background(), &types.CreateCompanyRequest{
			ProfileId: 1,
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("SetupProfiles", func(t *testing.T) {
		state.profileAID = createProfileHTTP(t, userA, emailA)
		state.profileBID = createProfileGRPC(t, userB, emailB)
	})

	t.Run("HTTPCreateCompany", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/companies", map[string]any{
			"name":            "Acme SRL",
			"registration_no": "REG-100",
			"fiscal_code":     "FISC-100",
			"profile_id":      state.profileAID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}
		var created types.CompanyResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create company failed: %v body=%s", err, string(body))
		}
		state.companyHTTPID = created.GetId()
	})

	t.Run("GRPCCreateCompany", func(t *testing.T) {
		created, err := grpcClient.CreateCompany(context.Background(), &types.CreateCompanyRequest{
			Name:           "Globex LLC",
			RegistrationNo: "REG-200",
			FiscalCode:     "FISC-200",
			ProfileId:      state.profileAID,
			Type:           "vendor",
		})
		if err != nil {
			t.Fatalf("grpc create company failed: %v", err)
		}
		state.companyGRPCID = created.GetId()
	})

	t.Run("HTTPGetByIDAfterGRPCCreate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/companies/"+strconv.FormatUint(state.companyGRPCID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
	})

	t.Run("GRPCGetByIDAfterHTTPCreate", func(t *testing.T) {
		got, err := grpcClient.GetCompany(context.Background(), &types.GetCompanyRequest{
			Id: state.companyHTTPID,
		})
		if err != nil {
			t.Fatalf("grpc get company failed: %v", err)
		}
		if got.GetName() != "Acme SRL" {
			t.Fatalf("unexpected company name: %+v", got)
		}
	})

	t.Run("HTTPListValidationMissingProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/companies", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListValidationBadProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/companies?profile_id=abc", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListByProfileID", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/companies?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var list types.ListCompaniesResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list companies failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 2 {
			t.Fatalf("expected at least 2 companies, got %d", list.GetTotal())
		}
		if !hasCompanyID(list.GetCompanies(), state.companyHTTPID) || !hasCompanyID(list.GetCompanies(), state.companyGRPCID) {
			t.Fatalf("expected created companies in list, got %+v", list.GetCompanies())
		}
	})

	t.Run("HTTPListByType", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/companies?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&type=vendor&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var list types.ListCompaniesResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list companies by type failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 1 {
			t.Fatalf("expected at least 1 company for type filter, got %d", list.GetTotal())
		}
		if !hasCompanyID(list.GetCompanies(), state.companyGRPCID) {
			t.Fatalf("expected company %d in filtered list", state.companyGRPCID)
		}
		for _, c := range list.GetCompanies() {
			if c.GetType() != "vendor" {
				t.Fatalf("expected only vendor companies, got type=%q", c.GetType())
			}
		}
	})

	t.Run("GRPCListByType", func(t *testing.T) {
		list, err := grpcClient.ListCompanies(context.Background(), &types.ListCompaniesRequest{
			ProfileId: state.profileAID,
			Type:      "vendor",
			Page:      1,
			PageSize:  10,
		})
		if err != nil {
			t.Fatalf("grpc list companies by type failed: %v", err)
		}
		if list.GetTotal() < 1 {
			t.Fatalf("expected at least 1 company for type filter, got %d", list.GetTotal())
		}
	})

	t.Run("HTTPCreateCompanySecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/companies", map[string]any{
			"name":            "Secondary SA",
			"registration_no": "REG-B",
			"fiscal_code":     "FISC-B",
			"profile_id":      state.profileBID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}
		var created types.CompanyResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create B failed: %v body=%s", err, string(body))
		}
		state.companyBID = created.GetId()
	})

	t.Run("HTTPListFilterBySecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/companies?profile_id="+strconv.FormatUint(state.profileBID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var list types.ListCompaniesResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list B failed: %v body=%s", err, string(body))
		}
		if !hasCompanyID(list.GetCompanies(), state.companyBID) {
			t.Fatalf("expected company %d in profile B list", state.companyBID)
		}
		for _, c := range list.GetCompanies() {
			if c.GetProfileId() != state.profileBID {
				t.Fatalf("expected only profile B companies, got profile_id=%d", c.GetProfileId())
			}
		}
	})

	t.Run("HTTPUpdate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodPut,
			"/companies/"+strconv.FormatUint(state.companyGRPCID, 10),
			map[string]any{
				"name":            "Globex Updated",
				"registration_no": "REG-200-U",
				"fiscal_code":     "FISC-200-U",
				"profile_id":      state.profileAID,
				"type":            "partner",
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
	})

	t.Run("GRPCGetAfterHTTPUpdate", func(t *testing.T) {
		got, err := grpcClient.GetCompany(context.Background(), &types.GetCompanyRequest{
			Id: state.companyGRPCID,
		})
		if err != nil {
			t.Fatalf("grpc get after update failed: %v", err)
		}
		if got.GetName() != "Globex Updated" || got.GetType() != "partner" {
			t.Fatalf("unexpected updated company payload: %+v", got)
		}
	})

	t.Run("GRPCDeleteThenHTTPNotFound", func(t *testing.T) {
		_, err := grpcClient.DeleteCompany(context.Background(), &types.DeleteCompanyRequest{
			Id: state.companyHTTPID,
		})
		if err != nil {
			t.Fatalf("grpc delete company failed: %v", err)
		}

		resp, _ := httpClient.doJSON(
			t,
			http.MethodGet,
			"/companies/"+strconv.FormatUint(state.companyHTTPID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
		}
	})
}

func hasCompanyID(companies []*types.CompanyResponse, id uint64) bool {
	for _, c := range companies {
		if c.GetId() == id {
			return true
		}
	}
	return false
}
