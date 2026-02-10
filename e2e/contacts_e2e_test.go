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

func TestContactsE2E_CrossTransportCRUDAndList(t *testing.T) {
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

		contactMinimalHTTPID uint64
		contactMinimalGRPCID uint64
		contactFullID        uint64
		contactBID           uint64
	}{}

	nowNanos := time.Now().UnixNano()
	userA := uint64(nowNanos%1_000_000) + 10_000
	userB := userA + 1
	emailA := fmt.Sprintf("contacts-e2e-a-%d@example.com", nowNanos)
	emailB := fmt.Sprintf("contacts-e2e-b-%d@example.com", nowNanos)

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
		if created.GetId() == 0 {
			t.Fatalf("expected profile id, got %+v", created)
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
		if created.GetId() == 0 {
			t.Fatalf("expected profile id, got %+v", created)
		}

		return created.GetId()
	}

	t.Run("HTTPCreateValidationMissingProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/contacts", map[string]any{})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPCreateValidationBadDOB", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodPost, "/contacts", map[string]any{
			"profile_id": 1,
			"dob":        "1990/01/02",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCCreateValidationMissingProfileID", func(t *testing.T) {
		_, err := grpcClient.CreateContact(context.Background(), &types.CreateContactRequest{})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("GRPCCreateValidationBadDOB", func(t *testing.T) {
		_, err := grpcClient.CreateContact(context.Background(), &types.CreateContactRequest{
			ProfileId: 1,
			Dob:       "1990/01/02",
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("SetupProfiles", func(t *testing.T) {
		state.profileAID = createProfileHTTP(t, userA, emailA)
		state.profileBID = createProfileGRPC(t, userB, emailB)
	})

	t.Run("HTTPCreateMinimalContact_OnlyRequiredField", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/contacts", map[string]any{
			"profile_id": state.profileAID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}

		var created types.ContactResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create contact failed: %v body=%s", err, string(body))
		}
		if created.GetId() == 0 || created.GetProfileId() != state.profileAID {
			t.Fatalf("unexpected created contact: %+v", created)
		}
		if created.GetFirstName() != "" || created.GetDob() != "" || created.GetType() != "" {
			t.Fatalf("expected optional fields to be empty by default, got %+v", created)
		}

		state.contactMinimalHTTPID = created.GetId()
	})

	t.Run("GRPCCreateMinimalContact_OnlyRequiredField", func(t *testing.T) {
		created, err := grpcClient.CreateContact(context.Background(), &types.CreateContactRequest{
			ProfileId: state.profileAID,
		})
		if err != nil {
			t.Fatalf("grpc create minimal contact failed: %v", err)
		}
		if created.GetId() == 0 || created.GetProfileId() != state.profileAID {
			t.Fatalf("unexpected created minimal contact: %+v", created)
		}
		if created.GetFirstName() != "" || created.GetLastName() != "" || created.GetNin() != "" || created.GetDob() != "" || created.GetPhone() != "" || created.GetType() != "" {
			t.Fatalf("expected all optional fields empty, got %+v", created)
		}

		state.contactMinimalGRPCID = created.GetId()
	})

	t.Run("GRPCCreateFullContact", func(t *testing.T) {
		created, err := grpcClient.CreateContact(context.Background(), &types.CreateContactRequest{
			ProfileId: state.profileAID,
			FirstName: "John",
			LastName:  "Doe",
			Nin:       "NIN-A-123",
			Dob:       "1990-01-02",
			Phone:     "+1-555-0101",
			Type:      "emergency",
		})
		if err != nil {
			t.Fatalf("grpc create contact failed: %v", err)
		}
		if created.GetId() == 0 || created.GetProfileId() != state.profileAID {
			t.Fatalf("unexpected created contact: %+v", created)
		}

		state.contactFullID = created.GetId()
	})

	t.Run("HTTPGetByIDAfterGRPCCreate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var got types.ContactResponse
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal get contact failed: %v body=%s", err, string(body))
		}
		if got.GetFirstName() != "John" || got.GetDob() != "1990-01-02" || got.GetType() != "emergency" {
			t.Fatalf("unexpected full contact payload: %+v", got)
		}
	})

	t.Run("GRPCGetByIDAfterHTTPCreate", func(t *testing.T) {
		got, err := grpcClient.GetContact(context.Background(), &types.GetContactRequest{
			Id: state.contactMinimalHTTPID,
		})
		if err != nil {
			t.Fatalf("grpc get contact failed: %v", err)
		}
		if got.GetProfileId() != state.profileAID {
			t.Fatalf("unexpected profile id, got %+v", got)
		}
		if got.GetFirstName() != "" || got.GetDob() != "" {
			t.Fatalf("expected optional empty fields, got %+v", got)
		}
	})

	t.Run("HTTPGetByIDAfterGRPCCreateMinimal", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts/"+strconv.FormatUint(state.contactMinimalGRPCID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var got types.ContactResponse
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal get minimal grpc contact failed: %v body=%s", err, string(body))
		}
		if got.GetFirstName() != "" || got.GetLastName() != "" || got.GetNin() != "" || got.GetDob() != "" || got.GetPhone() != "" || got.GetType() != "" {
			t.Fatalf("expected optional fields empty, got %+v", got)
		}
	})

	t.Run("HTTPUpdateWithOnlyRequiredField_ClearsOptionalFields", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodPut,
			"/contacts/"+strconv.FormatUint(state.contactMinimalHTTPID, 10),
			map[string]any{
				"profile_id": state.profileAID,
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var updated types.ContactResponse
		if err := json.Unmarshal(body, &updated); err != nil {
			t.Fatalf("unmarshal minimal update failed: %v body=%s", err, string(body))
		}
		if updated.GetFirstName() != "" || updated.GetLastName() != "" || updated.GetNin() != "" || updated.GetDob() != "" || updated.GetPhone() != "" || updated.GetType() != "" {
			t.Fatalf("expected optional fields empty after minimal update, got %+v", updated)
		}
	})

	t.Run("GRPCUpdateWithOnlyRequiredField_ClearsOptionalFields", func(t *testing.T) {
		updated, err := grpcClient.UpdateContact(context.Background(), &types.UpdateContactRequest{
			Id:        state.contactMinimalGRPCID,
			ProfileId: state.profileAID,
		})
		if err != nil {
			t.Fatalf("grpc minimal update failed: %v", err)
		}
		if updated.GetFirstName() != "" || updated.GetLastName() != "" || updated.GetNin() != "" || updated.GetDob() != "" || updated.GetPhone() != "" || updated.GetType() != "" {
			t.Fatalf("expected optional fields empty after grpc minimal update, got %+v", updated)
		}
	})

	t.Run("HTTPListValidationBadProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/contacts?profile_id=abc", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListValidationBadPage", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/contacts?page=abc", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListValidationPageSizeTooLarge", func(t *testing.T) {
		resp, _ := httpClient.doJSON(t, http.MethodGet, "/contacts?page_size=101", nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPListByProfileID", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var list types.ListContactsResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list contacts failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 2 {
			t.Fatalf("expected at least 2 contacts for profile A, got total=%d", list.GetTotal())
		}
		if !hasContactID(list.GetContacts(), state.contactMinimalHTTPID) ||
			!hasContactID(list.GetContacts(), state.contactMinimalGRPCID) ||
			!hasContactID(list.GetContacts(), state.contactFullID) {
			t.Fatalf("expected created contacts in list: %+v", list.GetContacts())
		}
	})

	t.Run("HTTPListByType", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&type=emergency&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var list types.ListContactsResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list contacts by type failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 1 {
			t.Fatalf("expected at least 1 contact for type filter, got total=%d", list.GetTotal())
		}
		if !hasContactID(list.GetContacts(), state.contactFullID) {
			t.Fatalf("expected contact %d in filtered list", state.contactFullID)
		}
		for _, c := range list.GetContacts() {
			if c.GetType() != "emergency" {
				t.Fatalf("expected only emergency contacts, got type=%q", c.GetType())
			}
		}
	})

	t.Run("HTTPListPagination", func(t *testing.T) {
		resp1, body1 := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=1&page_size=1",
			nil,
		)
		if resp1.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for page1, got %d body=%s", resp1.StatusCode, string(body1))
		}
		var page1 types.ListContactsResponse
		if err := json.Unmarshal(body1, &page1); err != nil {
			t.Fatalf("unmarshal page1 failed: %v", err)
		}
		if len(page1.GetContacts()) != 1 || page1.GetPage() != 1 || page1.GetPageSize() != 1 {
			t.Fatalf("unexpected page1 payload: %+v", page1)
		}

		resp2, body2 := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts?profile_id="+strconv.FormatUint(state.profileAID, 10)+"&page=2&page_size=1",
			nil,
		)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for page2, got %d body=%s", resp2.StatusCode, string(body2))
		}
		var page2 types.ListContactsResponse
		if err := json.Unmarshal(body2, &page2); err != nil {
			t.Fatalf("unmarshal page2 failed: %v", err)
		}
		if len(page2.GetContacts()) != 1 || page2.GetPage() != 2 || page2.GetPageSize() != 1 {
			t.Fatalf("unexpected page2 payload: %+v", page2)
		}
		if page1.GetTotal() != page2.GetTotal() {
			t.Fatalf("expected same total across pages, got page1=%d page2=%d", page1.GetTotal(), page2.GetTotal())
		}
	})

	t.Run("GRPCListByProfileID", func(t *testing.T) {
		list, err := grpcClient.ListContacts(context.Background(), &types.ListContactsRequest{
			ProfileId: state.profileAID,
			Page:      1,
			PageSize:  10,
		})
		if err != nil {
			t.Fatalf("grpc list contacts failed: %v", err)
		}
		if list.GetTotal() < 2 {
			t.Fatalf("expected at least 2 contacts for profile A, got total=%d", list.GetTotal())
		}
	})

	t.Run("GRPCListByType", func(t *testing.T) {
		list, err := grpcClient.ListContacts(context.Background(), &types.ListContactsRequest{
			ProfileId: state.profileAID,
			Type:      "emergency",
			Page:      1,
			PageSize:  10,
		})
		if err != nil {
			t.Fatalf("grpc list contacts by type failed: %v", err)
		}
		if list.GetTotal() < 1 {
			t.Fatalf("expected at least 1 contact for type filter, got total=%d", list.GetTotal())
		}
		for _, c := range list.GetContacts() {
			if c.GetType() != "emergency" {
				t.Fatalf("expected only emergency contacts, got type=%q", c.GetType())
			}
		}
	})

	t.Run("HTTPCreateContactOnSecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(t, http.MethodPost, "/contacts", map[string]any{
			"profile_id": state.profileBID,
			"first_name": "Secondary",
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, string(body))
		}

		var created types.ContactResponse
		if err := json.Unmarshal(body, &created); err != nil {
			t.Fatalf("unmarshal create contact B failed: %v body=%s", err, string(body))
		}
		state.contactBID = created.GetId()
	})

	t.Run("HTTPListFilterBySecondProfile", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts?profile_id="+strconv.FormatUint(state.profileBID, 10)+"&page=1&page_size=10",
			nil,
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var list types.ListContactsResponse
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("unmarshal list B failed: %v body=%s", err, string(body))
		}
		if list.GetTotal() < 1 {
			t.Fatalf("expected at least one contact for profile B, got total=%d", list.GetTotal())
		}
		if !hasContactID(list.GetContacts(), state.contactBID) {
			t.Fatalf("expected contact %d in profile B list", state.contactBID)
		}
		for _, c := range list.GetContacts() {
			if c.GetProfileId() != state.profileBID {
				t.Fatalf("expected only profile B contacts, got profile_id=%d", c.GetProfileId())
			}
		}
	})

	t.Run("HTTPUpdateValidationMissingProfileID", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodPut,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			map[string]any{},
		)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCUpdateValidationMissingProfileID", func(t *testing.T) {
		_, err := grpcClient.UpdateContact(context.Background(), &types.UpdateContactRequest{
			Id: state.contactFullID,
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("HTTPUpdateValidationBadDOB", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodPut,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			map[string]any{
				"profile_id": state.profileAID,
				"dob":        "01-02-1990",
			},
		)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPUpdate", func(t *testing.T) {
		resp, body := httpClient.doJSON(
			t,
			http.MethodPut,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			map[string]any{
				"profile_id": state.profileAID,
				"first_name": "Jane",
				"last_name":  "Smith",
				"nin":        "NIN-A-999",
				"dob":        "1988-05-20",
				"phone":      "+1-555-9999",
				"type":       "other",
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		var updated types.ContactResponse
		if err := json.Unmarshal(body, &updated); err != nil {
			t.Fatalf("unmarshal update contact failed: %v body=%s", err, string(body))
		}
		if updated.GetFirstName() != "Jane" || updated.GetDob() != "1988-05-20" || updated.GetType() != "other" {
			t.Fatalf("unexpected updated payload: %+v", updated)
		}
	})

	t.Run("GRPCGetAfterHTTPUpdate", func(t *testing.T) {
		got, err := grpcClient.GetContact(context.Background(), &types.GetContactRequest{
			Id: state.contactFullID,
		})
		if err != nil {
			t.Fatalf("grpc get contact after update failed: %v", err)
		}
		if got.GetFirstName() != "Jane" || got.GetDob() != "1988-05-20" {
			t.Fatalf("unexpected grpc updated contact: %+v", got)
		}
	})

	t.Run("GRPCDelete", func(t *testing.T) {
		_, err := grpcClient.DeleteContact(context.Background(), &types.DeleteContactRequest{
			Id: state.contactFullID,
		})
		if err != nil {
			t.Fatalf("grpc delete contact failed: %v", err)
		}
	})

	t.Run("HTTPDeleteAgainNotFound", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodDelete,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 after deleting already deleted contact, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTPGetByIDNotFound", func(t *testing.T) {
		resp, _ := httpClient.doJSON(
			t,
			http.MethodGet,
			"/contacts/"+strconv.FormatUint(state.contactFullID, 10),
			nil,
		)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for deleted contact, got %d", resp.StatusCode)
		}
	})

	t.Run("GRPCGetNotFound", func(t *testing.T) {
		_, err := grpcClient.GetContact(context.Background(), &types.GetContactRequest{Id: state.contactFullID})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected NotFound, got %v", err)
		}
	})

	t.Run("GRPCDeleteNotFound", func(t *testing.T) {
		_, err := grpcClient.DeleteContact(context.Background(), &types.DeleteContactRequest{Id: state.contactFullID})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected NotFound, got %v", err)
		}
	})
}

func hasContactID(contacts []*types.ContactResponse, id uint64) bool {
	for _, c := range contacts {
		if c.GetId() == id {
			return true
		}
	}
	return false
}
