package types

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateCompanyRequestValidate(t *testing.T) {
	valid := &CreateCompanyRequest{
		Name:           "ACME",
		RegistrationNo: "REG-1",
		FiscalCode:     "FISC-1",
		ProfileId:      7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if err := (&CreateCompanyRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewCreateCompanyRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/companies", strings.NewReader(`{"name":"ACME","registration_no":"REG-1","fiscal_code":"FISC-1","profile_id":7}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewCreateCompanyRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if parsed.GetName() != "ACME" || parsed.GetProfileId() != 7 {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}
}

func TestNewGetCompanyRequestFromContextAndValidate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/companies/10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	parsed, err := NewGetCompanyRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if err = parsed.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestUpdateCompanyRequestValidate(t *testing.T) {
	valid := &UpdateCompanyRequest{
		Id:             2,
		Name:           "ACME",
		RegistrationNo: "REG-1",
		FiscalCode:     "FISC-1",
		ProfileId:      7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if err := (&UpdateCompanyRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestListCompaniesRequestValidate(t *testing.T) {
	if err := (&ListCompaniesRequest{ProfileId: 1, PageSize: 100}).Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
	if err := (&ListCompaniesRequest{ProfileId: 0}).Validate(); err == nil {
		t.Fatal("expected validation error for missing profile_id")
	}
	if err := (&ListCompaniesRequest{ProfileId: 1, PageSize: 101}).Validate(); err == nil {
		t.Fatal("expected validation error for page_size > 100")
	}
}

func TestNewListCompaniesRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/companies?profile_id=7&page=2&page_size=30&type=vendor", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewListCompaniesRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if parsed.GetProfileId() != 7 || parsed.GetPage() != 2 || parsed.GetPageSize() != 30 || parsed.GetType() != "vendor" {
		t.Fatalf("unexpected parsed values: %+v", parsed)
	}
}
