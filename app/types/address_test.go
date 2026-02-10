package types

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateAddressRequestValidate(t *testing.T) {
	valid := &CreateAddressRequest{
		StreetName: "Street",
		StreenNo:   "10",
		City:       "City",
		County:     "County",
		Country:    "Country",
		ProfileId:  7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if err := (&CreateAddressRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewCreateAddressRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/addresses", strings.NewReader(`{"street_name":"Street","streen_no":"10","city":"City","county":"County","country":"Country","profile_id":7}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewCreateAddressRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if parsed.GetStreetName() != "Street" || parsed.GetProfileId() != 7 {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}
}

func TestNewGetAddressRequestFromContextAndValidate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/addresses/10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	parsed, err := NewGetAddressRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if err = parsed.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestUpdateAddressRequestValidate(t *testing.T) {
	valid := &UpdateAddressRequest{
		Id:         2,
		StreetName: "Street",
		StreenNo:   "10",
		City:       "City",
		County:     "County",
		Country:    "Country",
		ProfileId:  7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if err := (&UpdateAddressRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestListAddressesRequestValidate(t *testing.T) {
	if err := (&ListAddressesRequest{ProfileId: 1, PageSize: 100}).Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
	if err := (&ListAddressesRequest{ProfileId: 0}).Validate(); err == nil {
		t.Fatal("expected validation error for missing profile_id")
	}
	if err := (&ListAddressesRequest{ProfileId: 1, PageSize: 101}).Validate(); err == nil {
		t.Fatal("expected validation error for page_size > 100")
	}
}

func TestNewListAddressesRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/addresses?profile_id=7&page=2&page_size=30&type=billing", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewListAddressesRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if parsed.GetProfileId() != 7 || parsed.GetPage() != 2 || parsed.GetPageSize() != 30 || parsed.GetType() != "billing" {
		t.Fatalf("unexpected parsed values: %+v", parsed)
	}
}
