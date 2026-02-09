package types

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateProfileRequestValidate(t *testing.T) {
	valid := &CreateProfileRequest{UserId: 1, Email: "john@example.com"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	if err := (&CreateProfileRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error for missing fields")
	}
}

func TestNewCreateProfileRequestFromContext(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest("POST", "/profiles", strings.NewReader(`{"user_id":12,"email":"a@b.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	parsed, err := NewCreateProfileRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if parsed.GetUserId() != 12 || parsed.GetEmail() != "a@b.com" {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}

	badReq := httptest.NewRequest("POST", "/profiles", strings.NewReader("{invalid"))
	badReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	badCtx := e.NewContext(badReq, httptest.NewRecorder())
	if _, err := NewCreateProfileRequestFromContext(badCtx); err == nil {
		t.Fatal("expected bind error for invalid JSON")
	}
}

func TestNewGetProfileRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/profiles/10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("10")

	parsed, err := NewGetProfileRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if parsed.GetId() != 10 {
		t.Fatalf("expected id 10, got %d", parsed.GetId())
	}

	ctxInvalid := e.NewContext(req, httptest.NewRecorder())
	ctxInvalid.SetParamNames("id")
	ctxInvalid.SetParamValues("x")
	if _, err := NewGetProfileRequestFromContext(ctxInvalid); err == nil {
		t.Fatal("expected parse error for non-numeric id")
	}
}

func TestGetProfileRequestValidate(t *testing.T) {
	if err := (&GetProfileRequest{Id: 1}).Validate(); err != nil {
		t.Fatalf("expected valid request, got: %v", err)
	}
	if err := (&GetProfileRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error for id=0")
	}
}

func TestNewGetProfileByUserIDRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/profiles/user/22", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("user_id")
	ctx.SetParamValues("22")

	parsed, err := NewGetProfileByUserIDRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if parsed.GetUserId() != 22 {
		t.Fatalf("expected user_id 22, got %d", parsed.GetUserId())
	}
}

func TestGetProfileByUserIDValidate(t *testing.T) {
	if err := (&GetProfileByUserIDRequest{UserId: 1}).Validate(); err != nil {
		t.Fatalf("expected valid request, got: %v", err)
	}
	if err := (&GetProfileByUserIDRequest{}).Validate(); err == nil {
		t.Fatal("expected validation error for user_id=0")
	}
}

func TestNewUpdateProfileRequestFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("PUT", "/profiles/7", strings.NewReader(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("7")

	parsed, err := NewUpdateProfileRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if parsed.GetId() != 7 || parsed.GetEmail() != "new@example.com" {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}
}

func TestUpdateProfileRequestValidate(t *testing.T) {
	if err := (&UpdateProfileRequest{Id: 1, Email: "valid@example.com"}).Validate(); err != nil {
		t.Fatalf("expected valid request, got: %v", err)
	}
	if err := (&UpdateProfileRequest{Id: 0, Email: "a@b.com"}).Validate(); err == nil {
		t.Fatal("expected validation error for id=0")
	}
	if err := (&UpdateProfileRequest{Id: 1, Email: ""}).Validate(); err == nil {
		t.Fatal("expected validation error for empty email")
	}
}

func TestNewDeleteProfileRequestFromContextAndValidate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("DELETE", "/profiles/4", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("4")

	parsed, err := NewDeleteProfileRequestFromContext(ctx)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if err := parsed.Validate(); err != nil {
		t.Fatalf("expected valid request, got: %v", err)
	}

	if err := (&DeleteProfileRequest{Id: 0}).Validate(); err == nil {
		t.Fatal("expected validation error for id=0")
	}
}
