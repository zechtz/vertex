package services

import (
	"testing"

	"github.com/zechtz/vertex/internal/models"
)

func TestClassifyFailure(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string // expected code, "" for no diagnosis
	}{
		{
			// The line that actually took a user a support round trip to diagnose.
			// It arrives wrapped in Maven's colour codes.
			name: "lombok on a too-new jdk",
			line: "\x1b[1;31mERROR\x1b[m Failed to execute goal \x1b[32morg.apache.maven.plugins:" +
				"maven-compiler-plugin:3.10.1:compile\x1b[m on project \x1b[36mnest-uaa\x1b[m: " +
				"\x1b[1;31mFatal error compiling\x1b[m: java.lang.ExceptionInInitializerError: " +
				"com.sun.tools.javac.code.TypeTag :: UNKNOWN",
			want: "lombok_jdk_incompatible",
		},
		{
			name: "compiler crash without the lombok signature",
			line: "[ERROR] Failed to execute goal ... Fatal error compiling: something else",
			want: "compiler_crash",
		},
		{
			name: "dependency built for a newer jdk",
			line: "[ERROR] Unsupported class file major version 68",
			want: "class_version_unsupported",
		},
		{
			name: "jdk older than the target release",
			line: "[ERROR] invalid target release: 21",
			want: "jdk_too_old",
		},
		{
			name: "port conflict",
			line: "Web server failed to start. Port 8817 was already in use.",
			want: "port_in_use",
		},
		{
			name: "plain build failure",
			line: "\x1b[1;31mBUILD FAILURE\x1b[m",
			want: "build_failure",
		},
		{
			name: "ordinary build output is not a failure",
			line: "[INFO] --- maven-compiler-plugin:3.11.0:compile (default-compile) @ nest-gateway ---",
			want: "",
		},
		{
			name: "an application error mentioning a port is not a port conflict",
			line: "2026-09-09 16:46:49.800 ERROR --- CachingConnectionFactory: Shutdown Signal: channel error",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyFailure(tt.line)

			if tt.want == "" {
				if got != nil {
					t.Fatalf("classifyFailure() = %q, want no diagnosis", got.Code)
				}
				return
			}

			if got == nil {
				t.Fatalf("classifyFailure() = nil, want %q", tt.want)
			}
			if got.Code != tt.want {
				t.Errorf("code = %q, want %q", got.Code, tt.want)
			}
			if got.Suggestion == "" {
				t.Error("suggestion is empty; a diagnosis the user cannot act on is not useful")
			}
		})
	}
}

// The colour codes must never reach the UI, and must not prevent a match.
func TestClassifyFailureStripsAnsiFromDetail(t *testing.T) {
	reason := classifyFailure("\x1b[1;31mBUILD FAILURE\x1b[m")

	if reason == nil {
		t.Fatal("classifyFailure() = nil, want a diagnosis")
	}
	if reason.Detail != "BUILD FAILURE" {
		t.Errorf("detail = %q, want %q", reason.Detail, "BUILD FAILURE")
	}
}

// A build prints its specific cause before the generic "BUILD FAILURE" banner,
// so the first diagnosis must survive the rest of the output.
func TestRecordFailureReasonKeepsTheMostSpecific(t *testing.T) {
	service := &models.Service{ID: "svc", Name: "svc"}

	output := []string{
		"[INFO] Compiling 1832 source files",
		"[ERROR] Fatal error compiling: java.lang.ExceptionInInitializerError: com.sun.tools.javac.code.TypeTag :: UNKNOWN",
		"[INFO] BUILD FAILURE",
		"[ERROR] To see the full stack trace of the errors, re-run Maven with the -e switch.",
	}

	recorded := 0
	for _, line := range output {
		if reason := classifyFailure(line); reason != nil && recordFailureReason(service, reason) {
			recorded++
		}
	}

	if recorded != 1 {
		t.Errorf("recorded %d reasons, want 1 (only the first should publish)", recorded)
	}
	if service.FailureReason == nil {
		t.Fatal("no failure reason recorded")
	}
	if service.FailureReason.Code != "lombok_jdk_incompatible" {
		t.Errorf("code = %q, want the specific diagnosis, not the generic banner",
			service.FailureReason.Code)
	}
}
