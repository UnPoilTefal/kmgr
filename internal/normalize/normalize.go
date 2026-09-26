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

// ContextName is the canonical {user}@{cluster} identity kmgr assigns to a
// kubeconfig: it drives the source filename, the cluster name, and the
// AuthInfo name (namespaced by cluster to avoid collisions at merge time).
// It is distinct from the context key actually written inside a kubeconfig
// file, which may be overridden (see `kmgr import --ctx`).
type ContextName string

// New builds the canonical identity from raw user and cluster input,
// sanitizing both (see Name).
func New(user, cluster string) ContextName {
	return ContextName(Name(user) + "@" + Name(cluster))
}

// Parse reads an existing, already-conforming {user}@{cluster} string as-is,
// without sanitizing it. Use this for values a caller expects to already be
// valid (e.g. a context name typed on the command line), as opposed to New,
// which builds a fresh identity from untrusted parts.
func Parse(s string) (ContextName, bool) {
	at := strings.Index(s, "@")
	if at <= 0 || at >= len(s)-1 {
		return "", false
	}
	return ContextName(s), true
}

// FromFilename derives the canonical identity from a source kubeconfig
// filename. Expected format: kubeconfig_{user}@{cluster}.yaml.
func FromFilename(path string) (ContextName, bool) {
	base := strings.TrimSuffix(filepath.Base(path), ".yaml")
	if !strings.HasPrefix(base, "kubeconfig_") {
		return "", false
	}
	inner := strings.TrimPrefix(base, "kubeconfig_")
	at := strings.Index(inner, "@")
	if at <= 0 || at >= len(inner)-1 {
		return "", false
	}
	return New(inner[:at], inner[at+1:]), true
}

// String returns the identity as "{user}@{cluster}".
func (c ContextName) String() string { return string(c) }

// Cluster returns the cluster part of the identity.
func (c ContextName) Cluster() string {
	s := string(c)
	return s[strings.LastIndex(s, "@")+1:]
}

// User returns the user part of the identity (everything before the last "@").
func (c ContextName) User() string {
	s := string(c)
	return s[:strings.LastIndex(s, "@")]
}

// AuthInfo returns the AuthInfo name for this identity. It is namespaced by
// cluster (i.e. equal to the full identity) to avoid AuthInfo collisions
// when multiple source kubeconfigs are merged.
func (c ContextName) AuthInfo() string { return string(c) }

// SourceFilename returns the canonical "kubeconfig_{user}@{cluster}.yaml"
// filename for this identity.
func (c ContextName) SourceFilename() string {
	return "kubeconfig_" + string(c) + ".yaml"
}
