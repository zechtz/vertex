// Package services - discovery of the JDKs installed on this machine
package services

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// InstalledJDK is a JDK found on this machine. Path is what belongs in a
// JAVA_HOME, which is how a service is pinned to a particular JDK.
type InstalledJDK struct {
	Path         string `json:"path"`
	Version      string `json:"version"`
	MajorVersion int    `json:"majorVersion"`
	Source       string `json:"source"`
}

// jdkSearchRoot is a glob whose matches are candidate JAVA_HOME directories.
// Globbing where version managers actually install finds what is really there,
// unlike a fixed list of paths, which only finds the versions someone thought
// to write down.
type jdkSearchRoot struct {
	pattern string
	source  string
}

const jdkCacheTTL = 5 * time.Minute

var (
	jdkCacheMutex sync.Mutex
	jdkCache      []InstalledJDK
	jdkCachedAt   time.Time
)

// DiscoverJDKs lists the JDKs installed on this machine, newest major version
// last. Results are cached briefly: identifying each JDK means running it, and
// this is called to populate a picker, not in a hot path.
func DiscoverJDKs() []InstalledJDK {
	jdkCacheMutex.Lock()
	defer jdkCacheMutex.Unlock()

	if jdkCache != nil && time.Since(jdkCachedAt) < jdkCacheTTL {
		return jdkCache
	}

	jdkCache = discoverJDKs()
	jdkCachedAt = time.Now()

	return jdkCache
}

func discoverJDKs() []InstalledJDK {
	found := make([]InstalledJDK, 0)
	// Version managers expose the selected JDK as a symlink alongside the real
	// installs, so candidates are keyed by resolved path to list each JDK once.
	seen := make(map[string]bool)

	for _, root := range jdkSearchRoots() {
		matches, err := filepath.Glob(root.pattern)
		if err != nil {
			continue
		}

		for _, candidate := range matches {
			resolved, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				resolved = candidate
			}
			if seen[resolved] {
				continue
			}

			javaPath := filepath.Join(resolved, "bin", getJavaExecutable())
			if !isExecutable(javaPath) {
				continue
			}

			// Running it is the only trustworthy way to tell what a directory
			// holds; a broken or partial install fails here rather than being
			// offered to the user as a fix.
			version := getJavaVersion(javaPath)
			major := parseMajorJavaVersion(version)
			if major == 0 {
				continue
			}

			seen[resolved] = true
			found = append(found, InstalledJDK{
				Path:         resolved,
				Version:      version,
				MajorVersion: major,
				Source:       root.source,
			})
		}
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].MajorVersion != found[j].MajorVersion {
			return found[i].MajorVersion < found[j].MajorVersion
		}
		return found[i].Version < found[j].Version
	})

	log.Printf("[INFO] Discovered %d installed JDK(s)", len(found))
	return found
}

// jdkSearchRoots returns where JDKs are installed on this platform, covering the
// version managers first since a developer machine usually has several JDKs
// through one of them.
func jdkSearchRoots() []jdkSearchRoot {
	roots := make([]jdkSearchRoot, 0)

	if home := os.Getenv("HOME"); home != "" {
		roots = append(roots,
			jdkSearchRoot{filepath.Join(home, ".sdkman", "candidates", "java", "*"), "sdkman"},
			jdkSearchRoot{filepath.Join(home, ".asdf", "installs", "java", "*"), "asdf"},
			jdkSearchRoot{filepath.Join(home, ".jenv", "versions", "*"), "jenv"},
			jdkSearchRoot{filepath.Join(home, ".gradle", "jdks", "*"), "gradle"},
		)
	}

	switch runtime.GOOS {
	case "darwin":
		roots = append(roots,
			jdkSearchRoot{"/Library/Java/JavaVirtualMachines/*/Contents/Home", "system"},
			jdkSearchRoot{"/System/Library/Java/JavaVirtualMachines/*/Contents/Home", "system"},
			jdkSearchRoot{"/opt/homebrew/opt/openjdk*/libexec/openjdk.jdk/Contents/Home", "homebrew"},
			jdkSearchRoot{"/usr/local/opt/openjdk*/libexec/openjdk.jdk/Contents/Home", "homebrew"},
		)
	case "linux":
		roots = append(roots,
			jdkSearchRoot{"/usr/lib/jvm/*", "system"},
			jdkSearchRoot{"/opt/java/*", "system"},
		)
	case "windows":
		roots = append(roots,
			jdkSearchRoot{`C:\Program Files\Java\*`, "system"},
			jdkSearchRoot{`C:\Program Files\Eclipse Adoptium\*`, "system"},
			jdkSearchRoot{`C:\Program Files\Microsoft\jdk*`, "system"},
		)
	}

	// Whatever the machine is currently configured to use belongs in the list
	// even if it lives somewhere unusual.
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		roots = append(roots, jdkSearchRoot{javaHome, "JAVA_HOME"})
	}

	return roots
}

// parseMajorJavaVersion extracts the major version from a version string as
// reported by "java -version". It returns 0 when the string is not a version,
// which is how a directory that does not hold a usable JDK is rejected.
func parseMajorJavaVersion(version string) int {
	version = strings.TrimSpace(version)

	// Java 8 and earlier report "1.8.0_452", where the major version is the
	// second component; everything since reports "17.0.11" or "21".
	if strings.HasPrefix(version, "1.") {
		version = version[len("1."):]
	}

	digits := 0
	for digits < len(version) && version[digits] >= '0' && version[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0
	}

	major, err := strconv.Atoi(version[:digits])
	if err != nil {
		return 0
	}

	return major
}
