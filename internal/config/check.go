package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

// Issue describes a single problem found during an integrity check.
type Issue struct {
	Field string // "context", "cluster", "user", "permissions", "parse", "credentials", …
	Got   string
	Want  string // empty when there is no expected value (e.g. a parse error)
}

func (i Issue) String() string {
	if i.Want == "" {
		return fmt.Sprintf("%s: %s", i.Field, i.Got)
	}
	return fmt.Sprintf("%s: %q ≠ attendu %q", i.Field, i.Got, i.Want)
}

// ---------------------------------------------------------------------------
// Source files — file integrity only (no connectivity)
// ---------------------------------------------------------------------------

// SourceCheck holds the file-integrity result for one source kubeconfig.
type SourceCheck struct {
	File        string  // basename
	ContextName string  // context name found inside the file
	ClusterName string  // cluster name found inside the file
	UserName    string  // user name found inside the file
	Server      string  // server URL
	Issues      []Issue // naming convention + permission problems
}

func (c *SourceCheck) OK() bool { return len(c.Issues) == 0 }

// CheckSourceFiles runs file-integrity checks (naming convention, permissions)
// on every kubeconfig_*.yaml in configsDir. No connectivity is tested.
func CheckSourceFiles(configsDir string) ([]SourceCheck, error) {
	files, err := filepath.Glob(filepath.Join(configsDir, "kubeconfig_*.yaml"))
	if err != nil {
		return nil, err
	}
	results := make([]SourceCheck, len(files))
	for i, f := range files {
		results[i] = checkSourceFile(f)
	}
	return results, nil
}

func checkSourceFile(path string) SourceCheck {
	r := SourceCheck{File: filepath.Base(path)}

	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		r.Issues = append(r.Issues, Issue{Field: "parse", Got: err.Error()})
		return r
	}

	// Resolve primary context.
	ctxKey := cfg.CurrentContext
	if ctxKey == "" {
		for k := range cfg.Contexts {
			ctxKey = k
			break
		}
	}
	ctx := cfg.Contexts[ctxKey]
	r.ContextName = ctxKey
	if ctx != nil {
		r.ClusterName = ctx.Cluster
		r.UserName = ctx.AuthInfo
		if cl, ok := cfg.Clusters[ctx.Cluster]; ok {
			r.Server = cl.Server
		}
	}

	// Naming convention: context / cluster / user must match filename.
	expectedCtx := normalize.ContextFromFile(path)
	if ctxKey != expectedCtx {
		r.Issues = append(r.Issues, Issue{Field: "context", Got: ctxKey, Want: expectedCtx})
	}
	if ctx != nil && strings.Contains(expectedCtx, "@") {
		// AuthInfo est namespaced : deschampsf@ctain-d-00, identique au nom du contexte.
		// Le cluster est la partie après le dernier @.
		expectedCluster := expectedCtx[strings.LastIndex(expectedCtx, "@")+1:]
		if ctx.Cluster != expectedCluster {
			r.Issues = append(r.Issues, Issue{Field: "cluster", Got: ctx.Cluster, Want: expectedCluster})
		}
		if ctx.AuthInfo != expectedCtx {
			r.Issues = append(r.Issues, Issue{Field: "user", Got: ctx.AuthInfo, Want: expectedCtx})
		}
	}

	// Permissions must be 0600 — not applicable on Windows (ACL-based).
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(path)
		if err == nil && fi.Mode().Perm() != 0600 {
			r.Issues = append(r.Issues, Issue{
				Field: "permissions",
				Got:   fmt.Sprintf("%04o", fi.Mode().Perm()),
				Want:  "0600",
			})
		}
	}

	return r
}

// ---------------------------------------------------------------------------
// Target (merged) kubeconfig — structure integrity + connectivity per context
// ---------------------------------------------------------------------------

// ContextCheck holds the structural integrity and connectivity result for one
// context inside the merged kubeconfig.
type ContextCheck struct {
	ContextName   string
	ClusterName   string
	UserName      string
	Server        string
	Issues        []Issue // structural problems (missing cluster/user entry, no server, no credentials)
	Reachable     bool    // TCP+TLS OK (/version reachable)
	ReachErr      string  // error if not reachable
	Authenticated bool    // credentials accepted (/api/v1 returned 200)
	AuthErr       string  // error if not authenticated
}

func (c *ContextCheck) OK() bool {
	return len(c.Issues) == 0 && c.Reachable && c.Authenticated
}

// TargetCheck holds the full check result for the merged kubeconfig file.
type TargetCheck struct {
	Path     string
	Issues   []Issue        // file-level problems (parse, permissions)
	Contexts []ContextCheck // one per context declared in the merged file
}

func (t *TargetCheck) OK() bool {
	if len(t.Issues) > 0 {
		return false
	}
	for _, c := range t.Contexts {
		if !c.OK() {
			return false
		}
	}
	return true
}

// CheckTarget validates the structure of the merged kubeconfig and tests
// connectivity for every context it contains. Connectivity probes run in parallel.
func CheckTarget(mergedPath string) (TargetCheck, error) {
	result := TargetCheck{Path: mergedPath}

	fi, err := os.Stat(mergedPath)
	if os.IsNotExist(err) {
		result.Issues = append(result.Issues, Issue{Field: "fichier", Got: "introuvable — lance 'kcfg merge'"})
		return result, nil
	}
	if err != nil {
		return result, err
	}

	// Permissions — not applicable on Windows (ACL-based).
	if runtime.GOOS != "windows" && fi.Mode().Perm() != 0600 {
		result.Issues = append(result.Issues, Issue{
			Field: "permissions",
			Got:   fmt.Sprintf("%04o", fi.Mode().Perm()),
			Want:  "0600",
		})
	}

	cfg, err := clientcmd.LoadFromFile(mergedPath)
	if err != nil {
		result.Issues = append(result.Issues, Issue{Field: "parse", Got: err.Error()})
		return result, nil
	}

	// Build one ContextCheck per declared context; probe connectivity in parallel.
	type indexed struct {
		i int
		c ContextCheck
	}
	ch := make(chan indexed, len(cfg.Contexts))
	i := 0
	for name := range cfg.Contexts {
		go func(idx int, ctxName string) {
			ch <- indexed{idx, checkContext(mergedPath, ctxName)}
		}(i, name)
		i++
	}

	result.Contexts = make([]ContextCheck, len(cfg.Contexts))
	for range cfg.Contexts {
		item := <-ch
		result.Contexts[item.i] = item.c
	}

	return result, nil
}

// checkContext validates one context from the merged kubeconfig and probes connectivity.
func checkContext(mergedPath, ctxName string) ContextCheck {
	r := ContextCheck{ContextName: ctxName}

	// Re-load so we can use BuildConfigFromFlags with a context override cleanly.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = mergedPath
	overrides := &clientcmd.ConfigOverrides{CurrentContext: ctxName}
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	rawCfg, err := cc.RawConfig()
	if err != nil {
		r.Issues = append(r.Issues, Issue{Field: "parse", Got: err.Error()})
		return r
	}

	ctx, ok := rawCfg.Contexts[ctxName]
	if !ok || ctx == nil {
		r.Issues = append(r.Issues, Issue{Field: "context", Got: "entrée manquante dans le fichier mergé"})
		return r
	}
	r.ClusterName = ctx.Cluster
	r.UserName = ctx.AuthInfo

	// Cluster entry must exist and have a server URL.
	cl, hasCl := rawCfg.Clusters[ctx.Cluster]
	if !hasCl || cl == nil {
		r.Issues = append(r.Issues, Issue{Field: "cluster", Got: fmt.Sprintf("%q non trouvé dans le fichier mergé", ctx.Cluster)})
	} else {
		r.Server = cl.Server
		if cl.Server == "" {
			r.Issues = append(r.Issues, Issue{Field: "server", Got: "URL vide"})
		}
	}

	// User entry must exist and have at least one credential.
	u, hasU := rawCfg.AuthInfos[ctx.AuthInfo]
	if !hasU || u == nil {
		r.Issues = append(r.Issues, Issue{Field: "user", Got: fmt.Sprintf("%q non trouvé dans le fichier mergé", ctx.AuthInfo)})
	} else if u.Token == "" && u.ClientCertificateData == nil && u.ClientCertificate == "" && u.Exec == nil && u.AuthProvider == nil {
		r.Issues = append(r.Issues, Issue{Field: "credentials", Got: "aucune credential (token, cert, exec, auth-provider)"})
	}

	// Connectivity probes.
	restConfig, err := cc.ClientConfig()
	if err != nil {
		r.ReachErr = fmt.Sprintf("config: %v", err)
		return r
	}
	restConfig.Timeout = 5 * time.Second

	transport, err := rest.TransportFor(restConfig)
	if err != nil {
		r.ReachErr = fmt.Sprintf("transport: %v", err)
		return r
	}

	httpClient := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	// Sonde 1 — TCP+TLS : /version est public, valide juste l'accès réseau.
	resp, err := httpClient.Get(restConfig.Host + "/version")
	if err != nil {
		r.ReachErr = err.Error()
		return r
	}
	_ = resp.Body.Close()
	r.Reachable = true

	// Sonde 2 — authentification : /api/v1 exige des credentials valides.
	resp, err = httpClient.Get(restConfig.Host + "/api/v1")
	if err != nil {
		r.AuthErr = err.Error()
		return r
	}
	_ = resp.Body.Close()
	if resp.StatusCode == 200 {
		r.Authenticated = true
	} else {
		r.AuthErr = fmt.Sprintf("HTTP %d (credentials refusés ou insuffisants)", resp.StatusCode)
	}

	return r
}
