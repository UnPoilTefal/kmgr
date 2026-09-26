package config

import (
	"fmt"
	"strings"

	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

// OutputMode selects between the verbose human-oriented rendering and the
// compact machine-oriented rendering (see `kmgr --ai` / KMGR_AI).
type OutputMode int

const (
	Human OutputMode = iota
	AI
)

// Palette carries the ANSI color codes to use when rendering in Human mode.
// The zero value renders plain text (no color codes) — safe for tests.
type Palette struct {
	Reset, Red, Yellow, Green, Dim, Bold string
}

func orNone(s string) string {
	if s == "" {
		return "<aucun>"
	}
	return s
}

// Render renders one source-file check result: nothing in AI mode when the
// file is compliant, one line per issue otherwise.
func (s SourceCheck) Render(mode OutputMode, p Palette) string {
	if mode == AI {
		var b strings.Builder
		for _, issue := range s.Issues {
			fmt.Fprintf(&b, "source %s: %s\n", s.File, issue)
		}
		return b.String()
	}
	if s.OK() {
		return fmt.Sprintf("  %s✓%s %s\n", p.Green, p.Reset, s.File)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "  %s✗%s %s", p.Red, p.Reset, s.File)
	if s.Server != "" {
		fmt.Fprintf(&b, "  %s%s%s", p.Dim, s.Server, p.Reset)
	}
	b.WriteString("\n")
	for _, issue := range s.Issues {
		fmt.Fprintf(&b, "      %s⚠%s  %s\n", p.Yellow, p.Reset, issue)
	}
	return b.String()
}

// Render renders one context's structural + connectivity check result.
func (c ContextCheck) Render(mode OutputMode, p Palette) string {
	if mode == AI {
		state := "ok"
		switch {
		case len(c.Issues) > 0:
			state = fmt.Sprintf("invalid (%v)", c.Issues[0])
		case !c.Reachable:
			state = fmt.Sprintf("unreachable (%s)", c.ReachErr)
		case !c.Authenticated:
			state = fmt.Sprintf("auth failed (%s)", c.AuthErr)
		}
		return fmt.Sprintf("%s: %s\n", c.ContextName, state)
	}

	var b strings.Builder
	hasIssues := len(c.Issues) > 0
	switch {
	case !hasIssues && c.Reachable && c.Authenticated:
		fmt.Fprintf(&b, "  %s✓%s %s", p.Green, p.Reset, c.ContextName)
	case !hasIssues && c.Reachable && !c.Authenticated:
		fmt.Fprintf(&b, "  %s⚠%s %s", p.Yellow, p.Reset, c.ContextName)
	default:
		fmt.Fprintf(&b, "  %s✗%s %s", p.Red, p.Reset, c.ContextName)
	}
	if c.Server != "" {
		fmt.Fprintf(&b, "  %s%s%s", p.Dim, c.Server, p.Reset)
	}
	b.WriteString("\n")

	for _, issue := range c.Issues {
		fmt.Fprintf(&b, "      %s⚠%s  %s\n", p.Yellow, p.Reset, issue)
	}

	switch {
	case !c.Reachable:
		fmt.Fprintf(&b, "      %s✗%s  non joignable : %s%s%s\n", p.Red, p.Reset, p.Dim, c.ReachErr, p.Reset)
	case !c.Authenticated:
		fmt.Fprintf(&b, "      %s⚠%s  joignable — authentification échouée : %s%s%s\n", p.Yellow, p.Reset, p.Dim, c.AuthErr, p.Reset)
	default:
		fmt.Fprintf(&b, "      %s✓%s  joignable et authentifié\n", p.Green, p.Reset)
	}
	return b.String()
}

// StatusResult holds the active context and its connectivity, for `kmgr status`.
type StatusResult struct {
	ContextName  string
	ClusterName  string
	Server       string
	Connectivity ConnectivityResult
}

// Render renders the active context header and its connectivity result.
func (s StatusResult) Render(mode OutputMode, p Palette) string {
	var b strings.Builder
	if mode == AI {
		fmt.Fprintf(&b, "context: %s\nserver: %s\n", orNone(s.ContextName), orNone(s.Server))
	} else {
		fmt.Fprintf(&b, "  contexte : %s%s%s\n", p.Bold, orNone(s.ContextName), p.Reset)
		fmt.Fprintf(&b, "  cluster  : %s\n", orNone(s.ClusterName))
		fmt.Fprintf(&b, "  serveur  : %s%s%s\n", p.Dim, orNone(s.Server), p.Reset)
		b.WriteString("\n")
	}

	c := s.Connectivity
	switch {
	case !c.Reachable:
		if mode == AI {
			fmt.Fprintf(&b, "connectivity: unreachable (%v)\n", c.ReachErr)
		} else {
			fmt.Fprintf(&b, "%s✗%s  non joignable : %s%v%s\n", p.Red, p.Reset, p.Dim, c.ReachErr, p.Reset)
		}
	case !c.Authenticated:
		if mode == AI {
			fmt.Fprintf(&b, "connectivity: reachable, auth failed (%v)\n", c.AuthErr)
		} else {
			fmt.Fprintf(&b, "%s⚠%s  joignable — authentification échouée : %s%v%s\n", p.Yellow, p.Reset, p.Dim, c.AuthErr, p.Reset)
		}
	default:
		if mode == AI {
			b.WriteString("connectivity: ok\n")
		} else {
			fmt.Fprintf(&b, "%s✓%s joignable et authentifié\n", p.Green, p.Reset)
		}
	}
	return b.String()
}

// ManagedContext is one managed source kubeconfig, for `kmgr list`.
type ManagedContext struct {
	Identity normalize.ContextName
	File     string // basename
	Active   bool
}

// Render renders one managed context row.
func (m ManagedContext) Render(mode OutputMode, p Palette) string {
	if mode == AI {
		marker := "  "
		if m.Active {
			marker = "* "
		}
		return marker + m.Identity.String() + "\n"
	}
	if m.Active {
		return fmt.Sprintf("  %s✓%s %-40s %-20s %s%s%s\n", p.Green, p.Reset, m.Identity.String(), m.Identity.Cluster(), p.Dim, m.File, p.Reset)
	}
	return fmt.Sprintf("  %-40s %-20s %s\n", m.Identity.String(), m.Identity.Cluster(), m.File)
}
