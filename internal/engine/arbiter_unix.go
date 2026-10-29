//go:build unix

package engine

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Terminals do not report their initial focus state, so a screensaver that
// starts in a hidden tab (the usual case: the shell went idle *because* the
// user left) would render frames nobody sees. The arbiter makes sure at most
// one drift instance per user renders at a time: the instance that most
// recently started or regained focus owns rendering, the previous owner gets
// SIGUSR1 and pauses until its own terminal reports focus again.
//
// The owner file holds "pid starttime"; the start time (from /proc, empty
// elsewhere) guards against signalling an unrelated process that reused the pid.

// pauseSignal is the signal another instance sends to take over rendering.
var pauseSignal os.Signal = syscall.SIGUSR1

func ownerPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "drift-render-owner-"+strconv.Itoa(os.Getuid()))
}

// procStart returns the kernel start time of pid from /proc, or "" when
// unavailable (not Linux, process gone).
func procStart(pid int) string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return ""
	}
	// The command name in parentheses may contain spaces; fields after the
	// closing parenthesis are fixed, and starttime is the 22nd field overall.
	s := string(b)
	i := strings.LastIndexByte(s, ')')
	f := strings.Fields(s[i+1:])
	if len(f) < 20 {
		return ""
	}
	return f[19]
}

// readOwner returns the pid and start time recorded in the owner file, or 0.
func readOwner() (int, string) {
	b, err := os.ReadFile(ownerPath())
	if err != nil {
		return 0, ""
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0, ""
	}
	pid, _ := strconv.Atoi(f[0])
	start := ""
	if len(f) > 1 {
		start = f[1]
	}
	return pid, start
}

// isLive reports whether pid still is the process recorded with start.
func isLive(pid int, start string) bool {
	if pid <= 0 {
		return false
	}
	if start != "" {
		return procStart(pid) == start
	}
	return syscall.Kill(pid, 0) == nil
}

// claimRender makes this process the rendering owner and asks the previous
// owner, if any, to pause.
func claimRender() {
	me := os.Getpid()
	if prev, start := readOwner(); prev != 0 && prev != me && isLive(prev, start) {
		_ = syscall.Kill(prev, syscall.SIGUSR1)
	}
	tmp := ownerPath() + "." + strconv.Itoa(me)
	rec := strconv.Itoa(me) + " " + procStart(me) + "\n"
	if err := os.WriteFile(tmp, []byte(rec), 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, ownerPath())
}

// releaseRender removes the owner file if this process holds it.
func releaseRender() {
	if pid, _ := readOwner(); pid == os.Getpid() {
		_ = os.Remove(ownerPath())
	}
}
