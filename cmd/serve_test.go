package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	authclient "github.com/vibast-solutions/lib-go-auth/client"
	authmiddleware "github.com/vibast-solutions/lib-go-auth/middleware"
	authservice "github.com/vibast-solutions/lib-go-auth/service"
	"github.com/vibast-solutions/ms-go-profile/app/controller"
	"github.com/vibast-solutions/ms-go-profile/app/entity"
	"github.com/vibast-solutions/ms-go-profile/app/service"
)

type cmdRepoStub struct{}

func (cmdRepoStub) Create(context.Context, *entity.Profile) error             { return nil }
func (cmdRepoStub) FindByID(context.Context, uint64) (*entity.Profile, error) { return nil, nil }
func (cmdRepoStub) FindByUserID(context.Context, uint64) (*entity.Profile, error) {
	return nil, nil
}
func (cmdRepoStub) Update(context.Context, *entity.Profile) error { return nil }
func (cmdRepoStub) Delete(context.Context, uint64) error          { return nil }

type cmdContactRepoStub struct{}

func (cmdContactRepoStub) Create(context.Context, *entity.Contact) error             { return nil }
func (cmdContactRepoStub) FindByID(context.Context, uint64) (*entity.Contact, error) { return nil, nil }
func (cmdContactRepoStub) Update(context.Context, *entity.Contact) error             { return nil }
func (cmdContactRepoStub) Delete(context.Context, uint64) error                      { return nil }
func (cmdContactRepoStub) List(context.Context, uint64, string, uint32, uint32) ([]*entity.Contact, uint64, error) {
	return nil, 0, nil
}

type cmdAddressRepoStub struct{}

func (cmdAddressRepoStub) Create(context.Context, *entity.Address) error             { return nil }
func (cmdAddressRepoStub) FindByID(context.Context, uint64) (*entity.Address, error) { return nil, nil }
func (cmdAddressRepoStub) Update(context.Context, *entity.Address) error             { return nil }
func (cmdAddressRepoStub) Delete(context.Context, uint64) error                      { return nil }
func (cmdAddressRepoStub) List(context.Context, uint64, string, uint32, uint32) ([]*entity.Address, uint64, error) {
	return nil, 0, nil
}

type cmdCompanyRepoStub struct{}

func (cmdCompanyRepoStub) Create(context.Context, *entity.Company) error             { return nil }
func (cmdCompanyRepoStub) FindByID(context.Context, uint64) (*entity.Company, error) { return nil, nil }
func (cmdCompanyRepoStub) Update(context.Context, *entity.Company) error             { return nil }
func (cmdCompanyRepoStub) Delete(context.Context, uint64) error                      { return nil }
func (cmdCompanyRepoStub) List(context.Context, uint64, string, uint32, uint32) ([]*entity.Company, uint64, error) {
	return nil, 0, nil
}

type internalAuthClientStub struct{}

func (internalAuthClientStub) ValidateInternalAccess(_ context.Context, req authclient.InternalAccessRequest) (authclient.InternalAccessResponse, error) {
	switch req.APIKey {
	case "valid-key":
		return authclient.InternalAccessResponse{
			ServiceName:   "caller-service",
			AllowedAccess: []string{"profile-service"},
		}, nil
	case "no-access-key":
		return authclient.InternalAccessResponse{
			ServiceName:   "caller-service",
			AllowedAccess: []string{"notifications-service"},
		}, nil
	default:
		return authclient.InternalAccessResponse{}, &authclient.APIError{StatusCode: http.StatusUnauthorized}
	}
}

func newInternalAuthMiddlewareStub() *authmiddleware.EchoInternalAuthMiddleware {
	internalAuth := authservice.NewInternalAuthService(internalAuthClientStub{})
	return authmiddleware.NewEchoInternalAuthMiddleware(internalAuth)
}

func TestSetupHTTPServerHealthRoute(t *testing.T) {
	profileSvc := service.NewProfileService(cmdRepoStub{})
	profileCtrl := controller.NewProfileController(profileSvc)
	contactSvc := service.NewContactService(cmdContactRepoStub{})
	contactCtrl := controller.NewContactController(contactSvc)
	addressSvc := service.NewAddressService(cmdAddressRepoStub{})
	addressCtrl := controller.NewAddressController(addressSvc)
	companySvc := service.NewCompanyService(cmdCompanyRepoStub{})
	companyCtrl := controller.NewCompanyController(companySvc)
	internalAuthMW := newInternalAuthMiddlewareStub()
	e := setupHTTPServer(profileCtrl, contactCtrl, addressCtrl, companyCtrl, internalAuthMW, "profile-service")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestSetupHTTPServerHealthRouteForbidden(t *testing.T) {
	profileSvc := service.NewProfileService(cmdRepoStub{})
	profileCtrl := controller.NewProfileController(profileSvc)
	contactSvc := service.NewContactService(cmdContactRepoStub{})
	contactCtrl := controller.NewContactController(contactSvc)
	addressSvc := service.NewAddressService(cmdAddressRepoStub{})
	addressCtrl := controller.NewAddressController(addressSvc)
	companySvc := service.NewCompanyService(cmdCompanyRepoStub{})
	companyCtrl := controller.NewCompanyController(companySvc)
	internalAuthMW := newInternalAuthMiddlewareStub()
	e := setupHTTPServer(profileCtrl, contactCtrl, addressCtrl, companyCtrl, internalAuthMW, "profile-service")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-API-Key", "no-access-key")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestSetupHTTPServerHealthRouteAuthorized(t *testing.T) {
	profileSvc := service.NewProfileService(cmdRepoStub{})
	profileCtrl := controller.NewProfileController(profileSvc)
	contactSvc := service.NewContactService(cmdContactRepoStub{})
	contactCtrl := controller.NewContactController(contactSvc)
	addressSvc := service.NewAddressService(cmdAddressRepoStub{})
	addressCtrl := controller.NewAddressController(addressSvc)
	companySvc := service.NewCompanyService(cmdCompanyRepoStub{})
	companyCtrl := controller.NewCompanyController(companySvc)
	internalAuthMW := newInternalAuthMiddlewareStub()
	e := setupHTTPServer(profileCtrl, contactCtrl, addressCtrl, companyCtrl, internalAuthMW, "profile-service")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-API-Key", "valid-key")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health payload: %s", rec.Body.String())
	}
	if got := rec.Header().Get(echo.HeaderXRequestID); got == "" {
		t.Fatalf("expected %s header to be set", echo.HeaderXRequestID)
	}
}
