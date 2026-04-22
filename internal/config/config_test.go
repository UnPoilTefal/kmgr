package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

// writeKubeconfig writes a minimal valid kubeconfig to a temp file and returns its path.
func writeKubeconfig(t *testing.T, ctx, cluster, user, server string) string {
	t.Helper()
	content := "apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: " + server +
		"\n  name: " + cluster +
		"\ncontexts:\n- context:\n    cluster: " + cluster +
		"\n    user: " + user +
		"\n  name: " + ctx +
		"\ncurrent-context: " + ctx +
		"\nusers:\n- name: " + user +
		"\n  user:\n    token: fake-token\n"
	f, err := os.CreateTemp(t.TempDir(), "kubeconfig-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestNormalizeAndWrite(t *testing.T) {
	src := writeKubeconfig(t, "old-ctx", "old-cluster", "old-user", "https://k8s.example.com")
	dest := filepath.Join(t.TempDir(), "out.yaml")

	oldCtx, oldCluster, oldUser, err := config.NormalizeAndWrite(src, dest, "john@prod", "prod", "john@prod")
	if err != nil {
		t.Fatalf("NormalizeAndWrite error: %v", err)
	}
	if oldCtx != "old-ctx" {
		t.Errorf("oldCtx = %q, want %q", oldCtx, "old-ctx")
	}
	if oldCluster != "old-cluster" {
		t.Errorf("oldCluster = %q, want %q", oldCluster, "old-cluster")
	}
	if oldUser != "old-user" {
		t.Errorf("oldUser = %q, want %q", oldUser, "old-user")
	}

	// Dest file must exist.
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("dest file missing: %v", err)
	}
}

func TestCheckSourceFiles(t *testing.T) {
	dir := t.TempDir()

	// Conforming file.
	src := writeKubeconfig(t, "john@prod", "prod", "john@prod", "https://k8s.example.com")
	dest := filepath.Join(dir, "kubeconfig_john@prod.yaml")
	if err := os.Rename(src, dest); err != nil {
		t.Fatal(err)
	}

	results, err := config.CheckSourceFiles(dir)
	if err != nil {
		t.Fatalf("CheckSourceFiles error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	// Only permission issues are expected on some platforms; naming must be clean.
	for _, issue := range r.Issues {
		if issue.Field != "permissions" {
			t.Errorf("unexpected issue %v", issue)
		}
	}
}

func TestValidateKubeconfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		src := writeKubeconfig(t, "ctx", "cluster", "user", "https://k8s.example.com")
		if err := config.ValidateKubeconfig(src); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if err := config.ValidateKubeconfig("/does/not/exist.yaml"); err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		f, _ := os.CreateTemp(t.TempDir(), "bad-*.yaml")
		f.WriteString("not: valid: kubeconfig: [\n")
		f.Close()
		if err := config.ValidateKubeconfig(f.Name()); err == nil {
			t.Error("expected error for invalid yaml")
		}
	})
}
