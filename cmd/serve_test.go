package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
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
func (cmdContactRepoStub) List(context.Context, uint64, uint32, uint32) ([]*entity.Contact, uint64, error) {
	return nil, 0, nil
}

func TestSetupHTTPServerHealthRoute(t *testing.T) {
	profileSvc := service.NewProfileService(cmdRepoStub{})
	profileCtrl := controller.NewProfileController(profileSvc)
	contactSvc := service.NewContactService(cmdContactRepoStub{})
	contactCtrl := controller.NewContactController(contactSvc)
	e := setupHTTPServer(profileCtrl, contactCtrl)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
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
