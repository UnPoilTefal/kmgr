package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

func TestAddListRemoveMirror(t *testing.T) {
	t.Setenv("KMGR_DIR", t.TempDir())
	target := filepath.Join(t.TempDir(), "windows", "config")

	// Empty at start.
	mirrors, err := config.ListMirrors()
	if err != nil {
		t.Fatalf("ListMirrors error: %v", err)
	}
	if len(mirrors) != 0 {
		t.Fatalf("expected 0 mirrors, got %d", len(mirrors))
	}

	// Add.
	if err := config.AddMirror(target); err != nil {
		t.Fatalf("AddMirror error: %v", err)
	}
	mirrors, _ = config.ListMirrors()
	if len(mirrors) != 1 || mirrors[0] != target {
		t.Fatalf("expected [%s], got %v", target, mirrors)
	}

	// Duplicate must be rejected.
	if err := config.AddMirror(target); err == nil {
		t.Error("expected error on duplicate mirror")
	}

	// The merged file itself must be rejected.
	if err := config.AddMirror(config.MergedFile()); err == nil {
		t.Error("expected error when mirroring the merged file itself")
	}

	// Remove.
	if err := config.RemoveMirror(target); err != nil {
		t.Fatalf("RemoveMirror error: %v", err)
	}
	mirrors, _ = config.ListMirrors()
	if len(mirrors) != 0 {
		t.Fatalf("expected 0 mirrors after remove, got %v", mirrors)
	}

	// Removing an unknown mirror must fail.
	if err := config.RemoveMirror(target); err == nil {
		t.Error("expected error when removing unknown mirror")
	}
}

func TestSyncMirrors(t *testing.T) {
	t.Setenv("KMGR_DIR", t.TempDir())

	// Write a merged file.
	if err := os.WriteFile(config.MergedFile(), []byte("v1-content\n"), 0600); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "sub", "dir", "config")
	if err := config.AddMirror(target); err != nil {
		t.Fatal(err)
	}

	// First sync creates the target (including parent dirs).
	results := config.SyncMirrors()
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected sync results: %+v", results)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "v1-content\n" {
		t.Fatalf("target content = %q, err = %v", data, err)
	}
	if state, _ := config.CheckMirror(target); state != config.MirrorInSync {
		t.Errorf("state = %v, want in-sync", state)
	}

	// Update merged file → mirror becomes stale, sync refreshes it.
	if err := os.WriteFile(config.MergedFile(), []byte("v2-content\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if state, _ := config.CheckMirror(target); state != config.MirrorStale {
		t.Errorf("state = %v, want stale", state)
	}
	config.SyncMirrors()
	data, _ = os.ReadFile(target)
	if string(data) != "v2-content\n" {
		t.Errorf("target not refreshed: %q", data)
	}
}

func TestCheckMirrorMissing(t *testing.T) {
	t.Setenv("KMGR_DIR", t.TempDir())
	if err := os.WriteFile(config.MergedFile(), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	state, err := config.CheckMirror(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != config.MirrorMissing {
		t.Errorf("state = %v, want missing", state)
	}
}

func TestSyncMirrorsNoConfig(t *testing.T) {
	t.Setenv("KMGR_DIR", t.TempDir())
	if results := config.SyncMirrors(); results != nil {
		t.Errorf("expected nil results without mirrors, got %+v", results)
	}
}
