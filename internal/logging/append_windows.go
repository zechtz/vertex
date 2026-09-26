//go:build windows

package logging

import "os"

// isAppendOnlyFile is false on Windows. The Windows installer's batch file does
// not redirect Vertex's output to a file, so there is none to trim.
func isAppendOnlyFile(*os.File) bool {
	return false
}
