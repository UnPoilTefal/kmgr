package normalize

import (
	"path/filepath"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`[^a-z0-9@/_.-]`)

// Name lowercases s and replaces invalid characters with dashes.
func Name(s string) string {
	return re.ReplaceAllString(strings.ToLower(s), "-")
}

// ContextFromFile derives the canonical context name from a source filename.
// Expected format: kubeconfig_{user}@{cluster}.yaml → {user}@{cluster}
func ContextFromFile(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".yaml")
	base = strings.TrimPrefix(base, "kubeconfig_")
	parts := strings.SplitN(base, "@", 2)
	if len(parts) != 2 {
		return Name(base)
	}
	return Name(parts[0]) + "@" + Name(parts[1])
}

// IsValidSourceFilename returns true if the basename follows the
// kubeconfig_{user}@{cluster}.yaml convention (non-empty user and cluster).
func IsValidSourceFilename(path string) bool {
	base := strings.TrimSuffix(filepath.Base(path), ".yaml")
	if !strings.HasPrefix(base, "kubeconfig_") {
		return false
	}
	inner := strings.TrimPrefix(base, "kubeconfig_")
	at := strings.Index(inner, "@")
	return at > 0 && at < len(inner)-1
}
