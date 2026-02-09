package factory

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func TestNewModuleLoggerAddsModuleField(t *testing.T) {
	logger := NewModuleLogger("profile-controller")
	entry, ok := logger.(*logrus.Entry)
	if !ok {
		t.Fatalf("expected *logrus.Entry, got %T", logger)
	}

	got, ok := entry.Data["module"]
	if !ok || got != "profile-controller" {
		t.Fatalf("expected module field to be set, got: %#v", entry.Data)
	}
}

func TestLoggerWithContextAddsRequestID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := LoggerWithContext(logrus.NewEntry(logrus.New()), ctx)
	entry, ok := logger.(*logrus.Entry)
	if !ok {
		t.Fatalf("expected *logrus.Entry, got %T", logger)
	}

	got, ok := entry.Data["request_id"]
	if !ok || got != "req-123" {
		t.Fatalf("expected request_id field, got: %#v", entry.Data)
	}
}
