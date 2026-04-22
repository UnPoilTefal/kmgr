package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

// Dirs returns kubeDir, configsDir, backupDir.
func Dirs() (string, string, string) {
	base := os.Getenv("KCFG_DIR")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".kube")
	}
	return base, filepath.Join(base, "configs"), filepath.Join(base, "backups")
}

// MergedFile returns the path to the merged kubeconfig.
func MergedFile() string {
	kubeDir, _, _ := Dirs()
	return filepath.Join(kubeDir, "config")
}

// ValidateKubeconfig checks that path is a parseable kubeconfig with at least one context.
func ValidateKubeconfig(path string) error {
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return err
	}
	if len(cfg.Contexts) == 0 {
		return fmt.Errorf("aucun contexte trouvé")
	}
	return nil
}

// BackupFile creates a timestamped backup of a source kubeconfig file.
// Only the most recent backup is kept (limit = 1).
// Returns the backup path.
func BackupFile(path string) (string, error) {
	_, _, backupDir := Dirs()
	base := filepath.Base(path)
	bak, err := writeBackup(path, backupDir, base)
	if err != nil {
		return "", err
	}
	pruneBackups(backupDir, base, 1)
	return bak, nil
}

// BackupMerged creates a timestamped backup of the merged kubeconfig.
// Only the 5 most recent backups are kept.
// Returns the backup path, or "" if the merged file does not exist.
func BackupMerged() (string, error) {
	merged := MergedFile()
	if _, err := os.Stat(merged); os.IsNotExist(err) {
		return "", nil
	}
	_, _, backupDir := Dirs()
	bak, err := writeBackup(merged, backupDir, "config")
	if err != nil {
		return "", err
	}
	pruneBackups(backupDir, "config", 5)
	return bak, nil
}

// writeBackup writes a timestamped copy of src into backupDir.
// The backup filename is: {prefix}.{timestamp}.bak
func writeBackup(src, backupDir, prefix string) (string, error) {
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return "", err
	}
	bak := filepath.Join(backupDir, prefix+"."+time.Now().Format("20060102T150405")+".bak")
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return bak, os.WriteFile(bak, data, 0600)
}

// pruneBackups keeps only the `keep` most recent backups matching
// {backupDir}/{prefix}.*.bak, deleting the oldest ones.
func pruneBackups(backupDir, prefix string, keep int) {
	pattern := filepath.Join(backupDir, prefix+".*.bak")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) <= keep {
		return
	}
	// filepath.Glob returns sorted results; the timestamp format ensures
	// lexicographic order == chronological order.
	for _, old := range matches[:len(matches)-keep] {
		_ = os.Remove(old)
	}
}

// QuarantineFile moves a source kubeconfig file to configs/quarantine/.
// Returns the quarantine path.
func QuarantineFile(path string) (string, error) {
	_, configsDir, _ := Dirs()
	quarantineDir := filepath.Join(configsDir, "quarantine")
	if err := os.MkdirAll(quarantineDir, 0700); err != nil {
		return "", err
	}
	dest := filepath.Join(quarantineDir, filepath.Base(path))
	return dest, os.Rename(path, dest)
}

// NormalizeAndWrite loads a kubeconfig from srcPath, renames the primary
// context/cluster/user to the canonical names, and writes to destPath (0600).
// Returns the original context, cluster, and user names.
func NormalizeAndWrite(srcPath, destPath, ctxName, clusterName, userName string) (oldCtx, oldCluster, oldUser string, err error) {
	cfg, err := clientcmd.LoadFromFile(srcPath)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid kubeconfig: %w", err)
	}

	// Resolve the primary context key.
	primaryCtxKey := cfg.CurrentContext
	if primaryCtxKey == "" {
		for k := range cfg.Contexts {
			primaryCtxKey = k
			break
		}
	}
	if primaryCtxKey == "" {
		return "", "", "", fmt.Errorf("no context found in kubeconfig")
	}

	ctx := cfg.Contexts[primaryCtxKey]
	oldCtx = primaryCtxKey
	oldCluster = ctx.Cluster
	oldUser = ctx.AuthInfo

	// Build a fresh config with the renamed entries.
	newCfg := clientcmdapi.NewConfig()

	if c, ok := cfg.Clusters[oldCluster]; ok {
		newCfg.Clusters[clusterName] = c
	}
	if u, ok := cfg.AuthInfos[oldUser]; ok {
		newCfg.AuthInfos[userName] = u
	}

	newCtx := clientcmdapi.NewContext()
	newCtx.Cluster = clusterName
	newCtx.AuthInfo = userName
	newCtx.Namespace = ctx.Namespace
	newCfg.Contexts[ctxName] = newCtx
	newCfg.CurrentContext = ctxName

	if err := clientcmd.WriteToFile(*newCfg, destPath); err != nil {
		return "", "", "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0600); err != nil {
			return "", "", "", err
		}
	}
	return oldCtx, oldCluster, oldUser, nil
}

// MergeResult holds the outcome of a MergeAll call.
type MergeResult struct {
	Files          []string // merged source files (successfully parsed)
	Quarantined    []string // basenames moved to configs/quarantine/
	RestoredCtx    string   // context restored from before the merge, if any
	RestoredOnline bool     // whether the restored context was reachable (TCP+TLS)
	RestoredAuthed bool     // whether the restored context was authenticated
}

// MergeAll merges all kubeconfig_*.yaml files from configsDir into mergedPath.
// It preserves the current-context from the previous merged file when the
// context is still present in the new result.
func MergeAll(configsDir, mergedPath string) (*MergeResult, error) {
	files, err := filepath.Glob(filepath.Join(configsDir, "kubeconfig_*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	// Remember the current context before overwriting.
	previousCtx := CurrentContext()

	merged := clientcmdapi.NewConfig()
	result := &MergeResult{}
	for _, f := range files {
		// Gate 1 — naming convention: kubeconfig_{user}@{cluster}.yaml
		if !normalize.IsValidSourceFilename(f) {
			if _, err := QuarantineFile(f); err == nil {
				result.Quarantined = append(result.Quarantined, filepath.Base(f))
			}
			continue
		}
		// Gate 2 — parseability.
		cfg, err := clientcmd.LoadFromFile(f)
		if err != nil {
			if _, err := QuarantineFile(f); err == nil {
				result.Quarantined = append(result.Quarantined, filepath.Base(f))
			}
			continue
		}
		result.Files = append(result.Files, f)
		for k, v := range cfg.Clusters {
			merged.Clusters[k] = v
		}
		for k, v := range cfg.AuthInfos {
			merged.AuthInfos[k] = v
		}
		for k, v := range cfg.Contexts {
			merged.Contexts[k] = v
		}
	}

	// Restore current-context if the context still exists in the merged result.
	if previousCtx != "" {
		if _, exists := merged.Contexts[previousCtx]; exists {
			merged.CurrentContext = previousCtx
			result.RestoredCtx = previousCtx
			result.RestoredOnline, result.RestoredAuthed = probeContext(merged, previousCtx)
		}
	}

	if err := clientcmd.WriteToFile(*merged, mergedPath); err != nil {
		return nil, err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(mergedPath, 0600); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// probeContext tests connectivity and authentication for a given context.
// It writes the in-memory config to a temp file to avoid depending on the
// not-yet-written mergedPath.
// Returns (reachable, authenticated).
func probeContext(cfg *clientcmdapi.Config, ctxName string) (reachable, authenticated bool) {
	tmp, err := os.CreateTemp("", "kcfg-probe-*.yaml")
	if err != nil {
		return false, false
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	_ = tmp.Close()

	if err := clientcmd.WriteToFile(*cfg, tmp.Name()); err != nil {
		return false, false
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = tmp.Name()
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, &clientcmd.ConfigOverrides{CurrentContext: ctxName},
	)
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return false, false
	}

	restConfig.Timeout = 5 * time.Second

	transport, err := rest.TransportFor(restConfig)
	if err != nil {
		return false, false
	}
	httpClient := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	// Sonde 1 — TCP+TLS.
	resp, err := httpClient.Get(restConfig.Host + "/version")
	if err != nil {
		return false, false
	}
	_ = resp.Body.Close()

	// Sonde 2 — authentification.
	resp, err = httpClient.Get(restConfig.Host + "/api/v1")
	if err != nil {
		return true, false
	}
	_ = resp.Body.Close()
	return true, resp.StatusCode == 200
}

// ListContexts returns the sorted list of context names from the merged kubeconfig.
func ListContexts() []string {
	cfg, err := clientcmd.LoadFromFile(MergedFile())
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Contexts))
	for k := range cfg.Contexts {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// CurrentContext returns the active context from the merged kubeconfig.
func CurrentContext() string {
	cfg, err := clientcmd.LoadFromFile(MergedFile())
	if err != nil {
		return ""
	}
	return cfg.CurrentContext
}

// SetCurrentContext updates current-context in the merged kubeconfig.
func SetCurrentContext(ctxName string) error {
	mergedPath := MergedFile()
	cfg, err := clientcmd.LoadFromFile(mergedPath)
	if err != nil {
		return err
	}
	if _, ok := cfg.Contexts[ctxName]; !ok {
		return fmt.Errorf("context %q not found", ctxName)
	}
	cfg.CurrentContext = ctxName
	return clientcmd.WriteToFile(*cfg, mergedPath)
}

// ContextInfo returns ctxName, clusterName, and server URL of the active context.
func ContextInfo() (ctxName, clusterName, server string) {
	cfg, err := clientcmd.LoadFromFile(MergedFile())
	if err != nil {
		return "", "", ""
	}
	ctxName = cfg.CurrentContext
	if ctx, ok := cfg.Contexts[ctxName]; ok {
		clusterName = ctx.Cluster
		if cl, ok := cfg.Clusters[clusterName]; ok {
			server = cl.Server
		}
	}
	return
}

// ConnectivityResult holds the two-stage probe result for the active context.
type ConnectivityResult struct {
	Reachable     bool
	ReachErr      error
	Authenticated bool
	AuthErr       error
}

// TestConnectivity probes the active context in two stages:
// 1. TCP+TLS reachability via /version (public endpoint)
// 2. Authentication via /api/v1 (requires valid credentials)
func TestConnectivity() ConnectivityResult {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = MergedFile()
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return ConnectivityResult{ReachErr: err}
	}
	restConfig.Timeout = 5 * time.Second

	transport, err := rest.TransportFor(restConfig)
	if err != nil {
		return ConnectivityResult{ReachErr: err}
	}

	httpClient := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	// Sonde 1 — TCP+TLS.
	resp, err := httpClient.Get(restConfig.Host + "/version")
	if err != nil {
		return ConnectivityResult{ReachErr: err}
	}
	_ = resp.Body.Close()

	// Sonde 2 — authentification.
	resp, err = httpClient.Get(restConfig.Host + "/api/v1")
	if err != nil {
		return ConnectivityResult{Reachable: true, AuthErr: err}
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		return ConnectivityResult{
			Reachable: true,
			AuthErr:   fmt.Errorf("HTTP %d (credentials refusés ou insuffisants)", resp.StatusCode),
		}
	}
	return ConnectivityResult{Reachable: true, Authenticated: true}
}
