package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Liste les kubeconfigs gérés",
	RunE:  runList,
}

func runList(_ *cobra.Command, _ []string) error {
	section("Kubeconfigs gérés")

	_, configsDir, _ := config.Dirs()
	files, err := filepath.Glob(filepath.Join(configsDir, "kubeconfig_*.yaml"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		warn("Aucun kubeconfig importé. Lance : kcfg import -f <fichier> -u <user> -c <cluster>")
		return nil
	}

	currentCtx := config.CurrentContext()

	fmt.Printf("  %-40s %-20s %s\n", "CONTEXTE", "CLUSTER", "FICHIER")
	fmt.Printf("  %-40s %-20s %s\n", "--------", "-------", "-------")

	for _, f := range files {
		ctx := normalize.ContextFromFile(f)
		// ctx is "user@cluster" — take part after "@".
		cluster := ctx
		if at := strings.Index(ctx, "@"); at >= 0 {
			cluster = ctx[at+1:]
		}
		fname := filepath.Base(f)
		if ctx == currentCtx {
			fmt.Printf("  %s✓%s %-40s %-20s %s%s%s\n", green, reset, ctx, cluster, dim, fname, reset)
		} else {
			fmt.Printf("  %-40s %-20s %s\n", ctx, cluster, fname)
		}
	}

	fmt.Println()
	fmt.Printf("%sContexte actif : %s%s%s\n", dim, reset+bold, orNone(currentCtx), reset)
	return nil
}

func orNone(s string) string {
	if s == "" {
		return "<aucun>"
	}
	return s
}
