// Package services - classification of service startup failures
package services

import (
	"regexp"
	"strings"
	"time"

	"github.com/zechtz/vertex/internal/models"
)

// ansiEscape matches the colour codes Maven and Spring Boot emit. They are
// stripped before matching so a signature cannot be split by an escape
// sequence, and so the stored detail line is readable in the UI.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// failureSignature maps a marker in a service's output to an explanation of what
// went wrong and what to do about it.
type failureSignature struct {
	marker     string
	code       string
	summary    string
	suggestion string
}

// failureSignatures is ordered most specific first: the first match wins, so a
// precise diagnosis is preferred over the generic "BUILD FAILURE" that always
// accompanies it. Matching is case sensitive because these are fixed strings
// emitted by Maven, Gradle and the JVM.
var failureSignatures = []failureSignature{
	{
		marker:  "com.sun.tools.javac.code.TypeTag :: UNKNOWN",
		code:    "lombok_jdk_incompatible",
		summary: "Lombok cannot run on this JDK",
		suggestion: "This project's Lombok is older than the JDK compiling it. Set a JAVA_HOME " +
			"environment variable on this service pointing at an older JDK, or upgrade Lombok.",
	},
	{
		marker:  "Unsupported class file major version",
		code:    "class_version_unsupported",
		summary: "A dependency was built for a newer JDK than this service uses",
		suggestion: "Set a JAVA_HOME environment variable on this service pointing at a newer " +
			"JDK, or rebuild the dependency against this one.",
	},
	{
		marker:  "invalid target release",
		code:    "jdk_too_old",
		summary: "The JDK is older than the release this project targets",
		suggestion: "Set a JAVA_HOME environment variable on this service pointing at a JDK " +
			"matching the project's target release.",
	},
	{
		marker:  "was already in use",
		code:    "port_in_use",
		summary: "The service's port is already taken",
		suggestion: "Another process is listening on this port. Stop it, or change the port " +
			"configured for this service.",
	},
	{
		marker:  "Fatal error compiling",
		code:    "compiler_crash",
		summary: "The Java compiler failed while compiling this service",
		suggestion: "Usually an annotation processor that does not support the JDK in use. " +
			"Check the build output, and try a different JAVA_HOME for this service.",
	},
	{
		marker:     "COMPILATION ERROR",
		code:       "compile_error",
		summary:    "The service failed to compile",
		suggestion: "Check the build output for the offending source files.",
	},
	{
		marker:     "BUILD FAILURE",
		code:       "build_failure",
		summary:    "The build failed",
		suggestion: "Check the build output for the cause.",
	},
}

// classifyFailure reports what a line of service output says about a failure, or
// nil if the line carries no failure signal.
func classifyFailure(line string) *models.FailureReason {
	clean := ansiEscape.ReplaceAllString(line, "")

	for _, signature := range failureSignatures {
		if !strings.Contains(clean, signature.marker) {
			continue
		}

		return &models.FailureReason{
			Code:       signature.code,
			Summary:    signature.summary,
			Detail:     strings.TrimSpace(clean),
			Suggestion: signature.suggestion,
			DetectedAt: time.Now(),
		}
	}

	return nil
}

// recordFailureReason stores the first failure seen since the service last
// started. Later lines are ignored: a build prints its specific error before the
// generic "BUILD FAILURE" summary, so the first match is the most informative.
// It reports whether the reason was stored.
func recordFailureReason(service *models.Service, reason *models.FailureReason) bool {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	if service.FailureReason != nil {
		return false
	}

	service.FailureReason = reason
	return true
}
