package types

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateContactRequestValidate(t *testing.T) {
	valid := &CreateContactRequest{
		ProfileId: 5,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if err := (&CreateContactRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewCreateContactRequestFromContext(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest("POST", "/contacts", strings.NewReader(`{"profile_id":5}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewCreateContactRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if parsed.GetFirstName() != "" || parsed.GetProfileId() != 5 {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}
}

func TestNewGetContactRequestFromContextAndValidate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/contacts/10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	parsed, err := NewGetContactRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if err = parsed.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestUpdateContactRequestValidate(t *testing.T) {
	valid := &UpdateContactRequest{
		Id:        9,
		ProfileId: 5,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	badDOB := &UpdateContactRequest{
		Id:        9,
		Dob:       "02-01-1990",
		ProfileId: 5,
	}
	if err := badDOB.Validate(); err == nil {
		t.Fatal("expected validation error for bad dob")
	}
}

func TestNewListContactsRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/contacts?profile_id=5&page=2&page_size=30&type=emergency", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewListContactsRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if parsed.GetProfileId() != 5 || parsed.GetPage() != 2 || parsed.GetPageSize() != 30 || parsed.GetType() != "emergency" {
		t.Fatalf("unexpected parsed values: %+v", parsed)
	}
}

func TestListContactsRequestValidate(t *testing.T) {
	if err := (&ListContactsRequest{PageSize: 100}).Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
	if err := (&ListContactsRequest{PageSize: 101}).Validate(); err == nil {
		t.Fatal("expected validation error for page_size > 100")
	}
}
