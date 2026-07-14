package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// Mirrors — copies du kubeconfig mergé vers des chemins externes.
//
// Cas d'usage principal : sous WSL2, publier le kubeconfig côté Windows
// (/mnt/c/Users/<user>/.kube/config) pour que les outils Windows (Lens,
// kubectl.exe…) voient toujours la dernière version. kmgr étant le seul
// écrivain du fichier mergé, une copie one-way après chaque écriture suffit.
// ---------------------------------------------------------------------------

// MirrorsFile returns the path of the mirrors configuration file
// (one absolute target path per line, '#' starts a comment).
func MirrorsFile() string {
	kubeDir, _, _ := Dirs()
	return filepath.Join(kubeDir, "mirrors")
}

// ListMirrors returns the configured mirror target paths.
func ListMirrors() ([]string, error) {
	data, err := os.ReadFile(MirrorsFile())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var targets []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		targets = append(targets, line)
	}
	return targets, nil
}

// AddMirror registers a new mirror target.
func AddMirror(target string) error {
	if target == MergedFile() {
		return fmt.Errorf("le mirror ne peut pas être le fichier mergé lui-même : %s", target)
	}
	targets, err := ListMirrors()
	if err != nil {
		return err
	}
	for _, t := range targets {
		if t == target {
			return fmt.Errorf("mirror déjà configuré : %s", target)
		}
	}
	kubeDir, _, _ := Dirs()
	if err := os.MkdirAll(kubeDir, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(MirrorsFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = fmt.Fprintln(f, target)
	return err
}

// RemoveMirror unregisters a mirror target. The mirror file itself is not deleted.
func RemoveMirror(target string) error {
	targets, err := ListMirrors()
	if err != nil {
		return err
	}
	kept := make([]string, 0, len(targets))
	found := false
	for _, t := range targets {
		if t == target {
			found = true
			continue
		}
		kept = append(kept, t)
	}
	if !found {
		return fmt.Errorf("mirror non configuré : %s", target)
	}
	if len(kept) == 0 {
		return os.Remove(MirrorsFile())
	}
	return os.WriteFile(MirrorsFile(), []byte(strings.Join(kept, "\n")+"\n"), 0600)
}

// MirrorResult holds the outcome of one mirror synchronization.
type MirrorResult struct {
	Target string
	Err    error
}

// SyncMirrors copies the merged kubeconfig to every configured mirror target.
// Failures are per-target and never fatal for the caller.
func SyncMirrors() []MirrorResult {
	targets, err := ListMirrors()
	if err != nil || len(targets) == 0 {
		return nil
	}
	data, readErr := os.ReadFile(MergedFile())
	results := make([]MirrorResult, 0, len(targets))
	for _, t := range targets {
		r := MirrorResult{Target: t}
		switch {
		case readErr != nil:
			r.Err = fmt.Errorf("lecture du fichier mergé : %w", readErr)
		default:
			if err := os.MkdirAll(filepath.Dir(t), 0700); err != nil {
				r.Err = err
			} else if err := os.WriteFile(t, data, 0600); err != nil {
				// Sur drvfs (/mnt/c) le mode est ignoré, mais l'écriture passe.
				r.Err = err
			}
		}
		results = append(results, r)
	}
	return results
}

// MirrorState describes the freshness of one mirror target.
type MirrorState int

const (
	MirrorInSync  MirrorState = iota // identical to the merged file
	MirrorStale                      // exists but differs
	MirrorMissing                    // target does not exist
	MirrorError                      // unreadable target or merged file
)

func (s MirrorState) String() string {
	switch s {
	case MirrorInSync:
		return "in-sync"
	case MirrorStale:
		return "stale"
	case MirrorMissing:
		return "missing"
	default:
		return "error"
	}
}

// CheckMirror compares a mirror target with the merged kubeconfig.
func CheckMirror(target string) (MirrorState, error) {
	ref, err := os.ReadFile(MergedFile())
	if err != nil {
		return MirrorError, fmt.Errorf("lecture du fichier mergé : %w", err)
	}
	got, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return MirrorMissing, nil
	}
	if err != nil {
		return MirrorError, err
	}
	if bytes.Equal(ref, got) {
		return MirrorInSync, nil
	}
	return MirrorStale, nil
}

// DetectWindowsKubeconfig returns the Windows-side default kubeconfig path
// (%USERPROFILE%\.kube\config) translated to its WSL view (/mnt/c/Users/…).
// Only available when running inside WSL.
func DetectWindowsKubeconfig() (string, error) {
	if !IsWSL() {
		return "", fmt.Errorf("détection du profil Windows disponible uniquement sous WSL")
	}
	out, err := exec.Command("cmd.exe", "/C", "echo %USERPROFILE%").Output()
	if err != nil {
		// cmd.exe peut être absent du PATH ; PowerShell est le fallback.
		out, err = exec.Command("powershell.exe", "-NoProfile", "-Command", "$env:USERPROFILE").Output()
		if err != nil {
			return "", fmt.Errorf("impossible de récupérer %%USERPROFILE%% via cmd.exe/powershell.exe : %w", err)
		}
	}
	winProfile := strings.TrimSpace(string(out))
	if winProfile == "" {
		return "", fmt.Errorf("%%USERPROFILE%% est vide côté Windows")
	}
	wslPath, err := exec.Command("wslpath", "-u", winProfile).Output()
	if err != nil {
		return "", fmt.Errorf("conversion wslpath de %q : %w", winProfile, err)
	}
	return filepath.Join(strings.TrimSpace(string(wslPath)), ".kube", "config"), nil
}
