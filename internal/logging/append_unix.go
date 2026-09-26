//go:build !windows

package logging

import (
	"os"

	"golang.org/x/sys/unix"
)

// isAppendOnlyFile reports whether f is a regular file opened with O_APPEND.
func isAppendOnlyFile(f *os.File) bool {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	flags, err := unix.FcntlInt(f.Fd(), unix.F_GETFL, 0)
	return err == nil && flags&unix.O_APPEND != 0
}
