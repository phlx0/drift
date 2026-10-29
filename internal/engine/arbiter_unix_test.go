//go:build unix

package engine

import (
	"os"
	"testing"
)

func TestClaimAndReleaseRender(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	if pid, _ := readOwner(); pid != 0 {
		t.Fatalf("readOwner with no file = %d, want 0", pid)
	}

	claimRender()
	pid, start := readOwner()
	if pid != os.Getpid() {
		t.Fatalf("after claim owner = %d, want own pid %d", pid, os.Getpid())
	}
	if !isLive(pid, start) {
		t.Fatalf("own record (%d, %q) must be live", pid, start)
	}
	claimRender() // claiming again as the owner is a no-op
	if pid, _ := readOwner(); pid != os.Getpid() {
		t.Fatalf("after second claim owner = %d, want %d", pid, os.Getpid())
	}

	releaseRender()
	if _, err := os.Stat(ownerPath()); !os.IsNotExist(err) {
		t.Fatalf("release should remove the owner file, stat err = %v", err)
	}
}

func TestReleaseRenderKeepsAnotherOwner(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	// A file naming some other process must survive our release.
	if err := os.WriteFile(ownerPath(), []byte("999999999 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	releaseRender()
	b, err := os.ReadFile(ownerPath())
	if err != nil || string(b) != "999999999 1\n" {
		t.Fatalf("release removed or changed another owner's file: %q, %v", b, err)
	}
}

func TestIsLive(t *testing.T) {
	if isLive(0, "") || isLive(-1, "") {
		t.Fatal("isLive must be false for pid <= 0")
	}
	if !isLive(os.Getpid(), "") {
		t.Fatal("isLive(own pid) with no start time must be true")
	}
	if start := procStart(os.Getpid()); start != "" {
		if !isLive(os.Getpid(), start) {
			t.Fatal("isLive(own pid, own start) must be true")
		}
		if isLive(os.Getpid(), "not-a-start-time") {
			t.Fatal("isLive must reject a pid whose start time differs (pid reuse)")
		}
	}
}
