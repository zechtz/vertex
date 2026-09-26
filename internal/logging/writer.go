// Package logging shapes Vertex's own log: what the standard logger writes.
package logging

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

const (
	// maxFileSize is how large an append-only log file may grow before it is
	// trimmed, and keepSize how much of its end survives a trim.
	maxFileSize = 20 * 1024 * 1024
	keepSize    = 2 * 1024 * 1024
	// checkInterval is how many bytes are written between size checks, so the
	// file is not stat'ed on every line.
	checkInterval = 1024 * 1024
)

var debugTag = []byte("[DEBUG]")

// Writer is the standard logger's output. It drops [DEBUG] lines unless debug
// logging is on, and keeps the log file from growing without bound.
//
// Trimming applies only when out is a regular file opened for appending, which
// is how launchd hands Vertex its log file. After a truncation an appending
// descriptor writes at the new end of the file, so the trim is safe to do in
// place under the process that owns the descriptor. Anything else - a terminal,
// a pipe, journald - is written to untouched.
type Writer struct {
	mu         sync.Mutex
	out        *os.File
	debug      bool
	trimmable  bool
	maxSize    int64
	keepSize   int64
	sinceCheck int64
}

// NewWriter returns a Writer on out, trimming out at once if it is a log file
// that has already outgrown the limit.
func NewWriter(out *os.File, debug bool) *Writer {
	w := &Writer{
		out:       out,
		debug:     debug,
		trimmable: isAppendOnlyFile(out),
		maxSize:   maxFileSize,
		keepSize:  keepSize,
	}
	w.mu.Lock()
	w.trimIfTooLarge()
	w.mu.Unlock()
	return w
}

func (w *Writer) Write(p []byte) (int, error) {
	if !w.debug && isDebugLine(p) {
		return len(p), nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.out.Write(p)
	if w.trimmable {
		w.sinceCheck += int64(n)
		if w.sinceCheck >= checkInterval {
			w.sinceCheck = 0
			w.trimIfTooLarge()
		}
	}
	return n, err
}

// isDebugLine reports whether a log line is tagged [DEBUG], after whatever
// date and time prefix the logger's flags put in front of it.
func isDebugLine(line []byte) bool {
	message := bytes.TrimLeft(line, "0123456789/:. ")
	return bytes.HasPrefix(message, debugTag)
}

// dropDebugLines returns lines without the ones isDebugLine matches.
func dropDebugLines(lines []byte) []byte {
	var kept bytes.Buffer
	for _, line := range bytes.SplitAfter(lines, []byte("\n")) {
		if !isDebugLine(line) {
			kept.Write(line)
		}
	}
	return kept.Bytes()
}

// trimIfTooLarge cuts the file down to its last keepSize bytes, starting at a
// line boundary, once it passes maxSize. Callers must hold w.mu.
func (w *Writer) trimIfTooLarge() {
	if !w.trimmable {
		return
	}
	info, err := w.out.Stat()
	if err != nil || info.Size() <= w.maxSize {
		return
	}
	size := info.Size()

	// Keep the most recent lines when the descriptor can read them back; one
	// opened write-only is simply emptied.
	kept := make([]byte, w.keepSize)
	if n, err := w.out.ReadAt(kept, size-w.keepSize); err == nil && n == len(kept) {
		if newline := bytes.IndexByte(kept, '\n'); newline >= 0 {
			kept = kept[newline+1:]
		}
	} else {
		kept = nil
	}
	// The kept lines pass the same filter as new ones. On an upgrade that also
	// clears what older versions logged at [DEBUG] - environment variable
	// values among it - rather than carrying the tail of it forward.
	if !w.debug {
		kept = dropDebugLines(kept)
	}

	if err := w.out.Truncate(0); err != nil {
		return
	}
	fmt.Fprintf(w.out, "[INFO] Log trimmed by Vertex: kept %d KB from the end of %d MB\n",
		len(kept)/1024, size/(1024*1024))
	w.out.Write(kept)
}
