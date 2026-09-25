package services

import (
	"os"
	"strings"
	"testing"
)

// useTempCacheDir points os.UserCacheDir at a throwaway directory, so tests
// that write the Gradle init script leave the real cache alone.
func useTempCacheDir(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

// Gradle JVM options must reach the bootRun JVM. They used to be passed with
// --args, which hands them to the application as program arguments, so a
// service's -Xmx never took effect.
func TestGradleStartCommandPassesJavaOptsAsJVMArgs(t *testing.T) {
	useTempCacheDir(t)

	cmd, err := GetStartCommand(t.TempDir(), "gradle", "-Xmx512m -Dfoo=bar", "", false)
	if err != nil {
		t.Fatalf("GetStartCommand: %v", err)
	}

	if strings.Contains(cmd, "--args") {
		t.Errorf("command passes JVM options as program arguments: %s", cmd)
	}
	if !strings.Contains(cmd, `-PvertexAppJvmArgs="-Xmx512m -Dfoo=bar"`) {
		t.Errorf("command does not carry the JVM options to bootRun: %s", cmd)
	}

	initScript := initScriptPath(t, cmd)
	content, err := os.ReadFile(initScript)
	if err != nil {
		t.Fatalf("init script named in the command was not written: %v", err)
	}
	if string(content) != bootRunInitScript {
		t.Errorf("init script content = %q, want bootRunInitScript", content)
	}
}

// Without JVM options the Gradle command stays exactly as it was: no init
// script, no property.
func TestGradleStartCommandWithoutJavaOptsIsUnchanged(t *testing.T) {
	useTempCacheDir(t)

	dir := t.TempDir()
	cmd, err := GetStartCommand(dir, "gradle", "", "", false)
	if err != nil {
		t.Fatalf("GetStartCommand: %v", err)
	}

	if want := "cd " + dir + " && ./gradlew bootRun"; cmd != want {
		t.Errorf("command = %q, want %q", cmd, want)
	}
}

// initScriptPath extracts the path given to Gradle's -I flag.
func initScriptPath(t *testing.T, cmd string) string {
	t.Helper()

	_, rest, found := strings.Cut(cmd, `-I "`)
	if !found {
		t.Fatalf("command has no init script: %s", cmd)
	}
	path, _, _ := strings.Cut(rest, `"`)
	return path
}
