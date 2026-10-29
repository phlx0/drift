//go:build unix

package engine

import (
	"os"
	"strconv"
	"syscall"
)

// setNice lowers the scheduling priority of the current process by n
// (0 = unchanged). On Linux setpriority(2) is per-thread, so every existing
// thread is adjusted; threads created afterwards inherit the value. Errors are
// ignored: this is best-effort courtesy to the rest of the system, never a
// reason to fail.
func setNice(n int) {
	if n <= 0 {
		return
	}
	done := false
	if entries, err := os.ReadDir("/proc/self/task"); err == nil {
		for _, e := range entries {
			if tid, err := strconv.Atoi(e.Name()); err == nil {
				if syscall.Setpriority(syscall.PRIO_PROCESS, tid, n) == nil {
					done = true
				}
			}
		}
	}
	if !done {
		_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, n)
	}
}
