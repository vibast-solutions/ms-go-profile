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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestAddressE2E_CrossTransportCRUDAndList(t *testing.T) {
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

	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()
	grpcClient := types.NewProfileServiceClient(conn)

	state := struct {
		profileAID uint64
		profileBID uint64

		addressHTTPID uint64
		addressGRPCID uint64
		addressBID    uint64
	}{}

	nowNanos := time.Now().UnixNano()
	userA := uint64(nowNanos%1_000_000) + 20_000
	userB := userA + 1
	emailA := fmt.Sprintf("address-e2e-a-%d@example.com", nowNanos)
	emailB := fmt.Sprintf("address-e2e-b-%d@example.com", nowNanos)

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
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/addresses", map[string]any{
			"profile_id": 1,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCCreateValidationMissingMandatory", func(t *testing.T) {
		_, err := grpcClient.CreateAddress(context.Background(), &types.CreateAddressRequest{
			ProfileId: 1,
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("CreateValidationAdditionalDataTooLong", func(t *testing.T) {
		long := make([]byte, 513)
		for i := range long {
			long[i] = 'a'
		}
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/addresses", map[string]any{
			"street_name":     "Street",
			"streen_no":       "10",
			"city":            "City",
			"county":          "County",
			"country":         "Country",
			"profile_id":      1,
			"additional_data": string(long),
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("SetupProfiles", func(t *testing.T) {
		state.profileAID = createProfileHTTP(t, userA, emailA)
		state.profileBID = createProfileGRPC(t, userB, emailB)
	})

	t.Run("HTTPCreateAddress", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/addresses", map[string]any{
			"street_name": "Main Street",
			"streen_no":   "10",
			"city":        "London",
			"county":      "Greater London",
			"country":     "UK",
			"profile_id":  state.profileAID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}
		var created types.AddressResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create address failed: %v body=%s", err, string(body))
		}
		state.addressHTTPID = created.GetId()
	})

	t.Run("GRPCCreateAddress", func(t *testing.T) {
		created, err := grpcClient.CreateAddress(context.Background(), &types.CreateAddressRequest{
			StreetName:     "Second Street",
			StreenNo:       "20",
			City:           "Paris",
			County:         "Ile-de-France",
			Country:        "France",
			ProfileId:      state.profileAID,
			PostalCode:     "75000",
			Building:       "B",
			Apartment:      "12",
			AdditionalData: "entry code 1234",
			Type:           "billing",
		})
		if err != nil {
			t.Fatalf("grpc create address failed: %v", err)
		}
		state.addressGRPCID = created.GetId()
	})

	t.Run("HTTPGetByIDAfterGRPCCreate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/addresses/"+strconv.FormatUint(state.addressGRPCID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
	})

	t.Run("GRPCGetByIDAfterHTTPCreate", func(t *testing.T) {
		got, err := grpcClient.GetAddress(context.Background(), &types.GetAddressRequest{
			Id: state.addressHTTPID,
		})
		if err != nil {
			t.Fatalf("grpc get address failed: %v", err)
		}
		if got.GetStreetName() != "Main Street" {
			t.Fatalf("unexpected street_name: %+v", got)
		}
	})

	t.Run("HTTPListValidationMissingProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/addresses", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListValidationBadProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/addresses?profile_id=abc", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListByProfileID", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/addresses?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var list types.ListAddressesResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list addresses failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 2 {
			t.Fatalf("expected at least 2 addresses, got %d", list.GetTotal())
		}
		if !hasAddressID(list.GetAddresses(), state.addressHTTPID) || !hasAddressID(list.GetAddresses(), state.addressGRPCID) {
			t.Fatalf("expected both created addresses in list, got %+v", list.GetAddresses())
		}
	})

	t.Run("HTTPListPagination", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/addresses?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=1&page_size=1",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var page1 types.ListAddressesResponse
		if err := json.Unmarshal(body, &page1); err != nil {
			t.Fatalf("unmarshal list page1 failed: %v", err)
		}
		if len(page1.GetAddresses()) != 1 || page1.GetPage() != 1 || page1.GetPageSize() != 1 {
			t.Fatalf("unexpected page1 payload: %+v", page1)
		}
	})

	t.Run("GRPCListByProfileID", func(t *testing.T) {
		list, err := grpcClient.ListAddresses(context.Background(), &types.ListAddressesRequest{
			ProfileId: state.profileAID,
			Page:      1,
			PageSize:  10,
		})
		if err != nil {
			t.Fatalf("grpc list addresses failed: %v", err)
		}
		if list.GetTotal() < 2 {
			t.Fatalf("expected at least 2 addresses, got %d", list.GetTotal())
		}
	})

	t.Run("HTTPCreateAddressSecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/addresses", map[string]any{
			"street_name": "B Street",
			"streen_no":   "1",
			"city":        "Berlin",
			"county":      "Berlin",
			"country":     "Germany",
			"profile_id":  state.profileBID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}
		var created types.AddressResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create B failed: %v body=%s", err, string(body))
		}
		state.addressBID = created.GetId()
	})

	t.Run("HTTPListFilterBySecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/addresses?profile_id="+strconv.FormatUint(state.profileBID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
		var list types.ListAddressesResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list B failed: %v body=%s", err, string(body))
		}
		if !hasAddressID(list.GetAddresses(), state.addressBID) {
			t.Fatalf("expected address %d in profile B list", state.addressBID)
		}
		for _, a := range list.GetAddresses() {
			if a.GetProfileId() != state.profileBID {
				t.Fatalf("expected only profile B addresses, got profile_id=%d", a.GetProfileId())
			}
		}
	})

	t.Run("HTTPUpdateValidationMissingMandatory", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodPut,
			"/addresses/"+strconv.FormatUint(state.addressGRPCID, 10),
			map[string]any{"profile_id": state.profileAID},
		)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCUpdateValidationMissingMandatory", func(t *testing.T) {
		_, err := grpcClient.UpdateAddress(context.Background(), &types.UpdateAddressRequest{
			Id:        state.addressGRPCID,
			ProfileId: state.profileAID,
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("HTTPUpdate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodPut,
			"/addresses/"+strconv.FormatUint(state.addressGRPCID, 10),
			map[string]any{
				"street_name":     "Updated Street",
				"streen_no":       "99",
				"city":            "Madrid",
				"county":          "Madrid",
				"country":         "Spain",
				"profile_id":      state.profileAID,
				"postal_code":     "28001",
				"building":        "C",
				"apartment":       "7",
				"additional_data": "bell",
				"type":            "shipping",
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}
	})

	t.Run("GRPCGetAfterHTTPUpdate", func(t *testing.T) {
		got, err := grpcClient.GetAddress(context.Background(), &types.GetAddressRequest{
			Id: state.addressGRPCID,
		})
		if err != nil {
			t.Fatalf("grpc get after update failed: %v", err)
		}
		if got.GetStreetName() != "Updated Street" || got.GetCity() != "Madrid" {
			t.Fatalf("unexpected updated address: %+v", got)
		}
	})

	t.Run("GRPCDelete", func(t *testing.T) {
		_, err := grpcClient.DeleteAddress(context.Background(), &types.DeleteAddressRequest{
			Id: state.addressGRPCID,
		})
		if err != nil {
			t.Fatalf("grpc delete failed: %v", err)
		}
	})

	t.Run("HTTPDeleteAgainNotFound", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodDelete,
			"/addresses/"+strconv.FormatUint(state.addressGRPCID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCGetNotFound", func(t *testing.T) {
		_, err := grpcClient.GetAddress(context.Background(), &types.GetAddressRequest{Id: state.addressGRPCID})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected NotFound, got %v", err)
		}
	})
}

func hasAddressID(addresses []*types.AddressResponse, id uint64) bool {
	for _, a := range addresses {
		if a.GetId() == id {
			return true
		}
	}
	return false
}
