package logging

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openLogFile opens a file the way launchd opens StandardErrorPath: for
// reading and appending.
func openLogFile(t *testing.T, flags int) *os.File {
	t.Helper()

	f, err := os.OpenFile(filepath.Join(t.TempDir(), "vertex.stderr.log"), flags|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func readAll(t *testing.T, f *os.File) string {
	t.Helper()

	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

// numberedLines writes count lines of about 100 bytes, "line 0" to "line
// count-1", so a trim can be checked for where it cut.
func numberedLines(count int) []byte {
	var buf bytes.Buffer
	for i := 0; i < count; i++ {
		fmt.Fprintf(&buf, "2026/09/26 06:00:00 [INFO] line %d %s\n", i, strings.Repeat("x", 60))
	}
	return buf.Bytes()
}

func TestDebugLinesAreDroppedUnlessDebugIsOn(t *testing.T) {
	lines := []string{
		"2026/09/26 06:00:00 [DEBUG] Eureka app[0]: NEST-GATEWAY with 1 instances\n",
		"2026/09/26 06:00:00.123456 [DEBUG] with microseconds\n",
		"[DEBUG] with no timestamp\n",
		"2026/09/26 06:00:00 [INFO] Service gateway started\n",
		"2026/09/26 06:00:00 Warning: message mentioning [DEBUG] later on\n",
	}

	for _, debug := range []bool{false, true} {
		f := openLogFile(t, os.O_RDWR|os.O_APPEND)
		w := NewWriter(f, debug)
		for _, line := range lines {
			if n, err := w.Write([]byte(line)); err != nil || n != len(line) {
				t.Fatalf("Write() = %d, %v; want %d, nil", n, err, len(line))
			}
		}

		got := readAll(t, f)
		for i, line := range lines {
			isDebug := i < 3
			if kept := strings.Contains(got, line); kept != (debug || !isDebug) {
				t.Errorf("debug=%v: line %q kept = %v", debug, strings.TrimSpace(line), kept)
			}
		}
	}
}

// An existing install's log is trimmed on the first start after upgrading,
// keeping its most recent lines from a line boundary on.
func TestOversizedLogIsTrimmedAtStart(t *testing.T) {
	f := openLogFile(t, os.O_RDWR|os.O_APPEND)
	original := numberedLines(30000) // about 3 MB
	if _, err := f.Write(original); err != nil {
		t.Fatal(err)
	}

	w := &Writer{out: f, debug: true, trimmable: isAppendOnlyFile(f), maxSize: 1024 * 1024, keepSize: 64 * 1024}
	if !w.trimmable {
		t.Fatal("an O_APPEND regular file should be trimmable")
	}
	w.trimIfTooLarge()

	got := readAll(t, f)
	if int64(len(got)) > w.keepSize+200 {
		t.Fatalf("trimmed log is %d bytes, want about %d", len(got), w.keepSize)
	}
	if !strings.HasPrefix(got, "[INFO] Log trimmed by Vertex") {
		t.Errorf("trimmed log does not open with the trim notice: %q", got[:80])
	}
	body := strings.SplitN(got, "\n", 2)[1]
	if !strings.HasPrefix(body, "2026/09/26 06:00:00 [INFO] line ") {
		t.Errorf("kept lines do not start at a line boundary: %q", body[:80])
	}
	if !strings.HasSuffix(got, "line 29999 "+strings.Repeat("x", 60)+"\n") {
		t.Errorf("the most recent line was not kept")
	}

	// Writes after the trim land at the new end, with no gap before them.
	if _, err := w.Write([]byte("after the trim\n")); err != nil {
		t.Fatal(err)
	}
	if after := readAll(t, f); after != got+"after the trim\n" {
		t.Errorf("write after trim did not append cleanly; file is %d bytes, want %d", len(after), len(got)+15)
	}
}

// A running Vertex trims as it writes, not only at start.
func TestLogIsTrimmedWhileWriting(t *testing.T) {
	f := openLogFile(t, os.O_RDWR|os.O_APPEND)
	w := NewWriter(f, true)
	w.maxSize, w.keepSize = 512*1024, 32*1024

	for _, line := range bytes.SplitAfter(numberedLines(20000), []byte("\n")) {
		w.Write(line)
	}

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if limit := w.maxSize + checkInterval; info.Size() > limit {
		t.Errorf("log reached %d bytes while writing, want at most %d", info.Size(), limit)
	}
}

// A file Vertex was not handed to append to - here one opened write-only at
// an offset - is left alone: truncating it would leave the next write past
// the end of an empty file.
func TestFileWithoutAppendIsNotTrimmed(t *testing.T) {
	f := openLogFile(t, os.O_WRONLY)
	if _, err := f.Write(numberedLines(30000)); err != nil {
		t.Fatal(err)
	}

	w := &Writer{out: f, trimmable: isAppendOnlyFile(f), maxSize: 1024, keepSize: 512}
	if w.trimmable {
		t.Fatal("a file opened without O_APPEND must not be trimmable")
	}
	w.trimIfTooLarge()

	if size := len(readAll(t, f)); size < 1024*1024 {
		t.Errorf("file shrank to %d bytes, want it untouched", size)
	}
}

// Older versions logged environment variable values, always at [DEBUG]. With
// debug logging off, the trim drops [DEBUG] lines from what it keeps, so an
// upgrade does not carry those values forward.
func TestTrimDropsDebugLinesFromWhatItKeeps(t *testing.T) {
	f := openLogFile(t, os.O_RDWR|os.O_APPEND)
	var old bytes.Buffer
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&old, "2026/09/25 10:00:00 [DEBUG]   CONFIG_SERVER_PASSWORD=hunter2-%d\n", i)
		fmt.Fprintf(&old, "2026/09/25 10:00:00 [INFO] Service gateway started %d\n", i)
	}
	if _, err := f.Write(old.Bytes()); err != nil {
		t.Fatal(err)
	}

	w := &Writer{out: f, trimmable: isAppendOnlyFile(f), maxSize: 256 * 1024, keepSize: 64 * 1024}
	w.trimIfTooLarge()

	got := readAll(t, f)
	if strings.Contains(got, "hunter2") || strings.Contains(got, "[DEBUG]") {
		t.Errorf("trimmed log still holds [DEBUG] lines")
	}
	if !strings.Contains(got, "[INFO] Service gateway started 19999\n") {
		t.Errorf("trimmed log lost the most recent [INFO] line")
	}
}
