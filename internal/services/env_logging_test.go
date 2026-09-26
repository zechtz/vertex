package services

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// captureLog collects what the standard logger writes while fn runs.
func captureLog(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previous)

	fn()
	return buf.String()
}

// The start-up environment is logged by name only. Its values can be
// credentials, and the log is a plain file kept for weeks.
func TestStartupEnvIsLoggedByNameOnly(t *testing.T) {
	output := captureLog(t, func() {
		logStartupEnvNames("gateway", []string{
			"CONFIG_SERVER_PASSWORD=hunter2",
			"SPRING_PROFILES_ACTIVE=dev",
			"JAVA_HOME=/opt/jdk-21",
			"DB_PASSWORD=not-a-startup-var",
		})
	})

	for _, value := range []string{"hunter2", "not-a-startup-var", "=dev", "/opt/jdk-21"} {
		if strings.Contains(output, value) {
			t.Errorf("log contains the value %q:\n%s", value, output)
		}
	}
	for _, name := range []string{"CONFIG_SERVER_PASSWORD", "SPRING_PROFILES_ACTIVE", "JAVA_HOME"} {
		if !strings.Contains(output, name) {
			t.Errorf("log does not name %s:\n%s", name, output)
		}
	}
	if strings.Contains(output, "DB_PASSWORD") {
		t.Errorf("log names a variable outside the start-up set:\n%s", output)
	}
}
