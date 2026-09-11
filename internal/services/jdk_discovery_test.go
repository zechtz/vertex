package services

import "testing"

func TestParseMajorJavaVersion(t *testing.T) {
	tests := []struct {
		version string
		want    int
	}{
		{"17.0.11", 17},
		{"21", 21},
		{"25-ea", 25},
		{"1.8.0_452", 8}, // Java 8 and earlier report a 1.x prefix
		{"11.0.23+9", 11},
		{"unknown", 0}, // what getJavaVersion returns for a directory that is not a JDK
		{"", 0},
		{"not-a-version", 0},
	}

	for _, tt := range tests {
		if got := parseMajorJavaVersion(tt.version); got != tt.want {
			t.Errorf("parseMajorJavaVersion(%q) = %d, want %d", tt.version, got, tt.want)
		}
	}
}

// Discovery must never offer a path that cannot serve as a JAVA_HOME, since the
// point of the list is to be applied directly to a service.
func TestDiscoverJDKsReturnsUsableEntries(t *testing.T) {
	jdks := DiscoverJDKs()

	if len(jdks) == 0 {
		t.Skip("no JDKs installed on this machine")
	}

	seen := make(map[string]bool)
	for _, jdk := range jdks {
		if jdk.Path == "" {
			t.Error("JDK has no path; it could not be used as JAVA_HOME")
		}
		if jdk.MajorVersion <= 0 {
			t.Errorf("JDK at %s has major version %d, want a positive version",
				jdk.Path, jdk.MajorVersion)
		}
		if jdk.Source == "" {
			t.Errorf("JDK at %s has no source", jdk.Path)
		}
		if seen[jdk.Path] {
			t.Errorf("JDK at %s listed more than once; symlinked installs should collapse", jdk.Path)
		}
		seen[jdk.Path] = true
	}

	// Ascending major version, so a picker reads in a predictable order.
	for i := 1; i < len(jdks); i++ {
		if jdks[i].MajorVersion < jdks[i-1].MajorVersion {
			t.Errorf("JDKs out of order: %d before %d",
				jdks[i-1].MajorVersion, jdks[i].MajorVersion)
		}
	}

	t.Logf("discovered %d JDK(s)", len(jdks))
	for _, jdk := range jdks {
		t.Logf("  Java %d (%s) via %s at %s", jdk.MajorVersion, jdk.Version, jdk.Source, jdk.Path)
	}
}
