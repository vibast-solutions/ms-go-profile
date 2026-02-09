package cmd

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/vibast-solutions/ms-go-profile/config"
)

func TestConfigureLoggingValidLevel(t *testing.T) {
	previousLevel := logrus.GetLevel()
	previousFormatter := logrus.StandardLogger().Formatter
	t.Cleanup(func() {
		logrus.SetLevel(previousLevel)
		logrus.SetFormatter(previousFormatter)
	})

	err := configureLogging(&config.Config{LogLevel: "debug"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Fatalf("expected debug level, got %s", logrus.GetLevel())
	}
}

func TestConfigureLoggingInvalidLevel(t *testing.T) {
	err := configureLogging(&config.Config{LogLevel: "definitely-not-a-level"})
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}
