package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Affiche le contexte actif et teste la connexion",
	RunE:  runStatus,
}

func runStatus(_ *cobra.Command, _ []string) error {
	warnEnvDesync()

	section("Status")

	ctxName, clusterName, server := config.ContextInfo()
	fmt.Printf("  contexte : %s%s%s\n", bold, orNone(ctxName), reset)
	fmt.Printf("  cluster  : %s\n", orNone(clusterName))
	fmt.Printf("  serveur  : %s%s%s\n", dim, orNone(server), reset)
	fmt.Println()

	result := config.TestConnectivity()
	switch {
	case !result.Reachable:
		fmt.Fprintf(os.Stderr, "%s✗%s  non joignable : %s%v%s\n", red, reset, dim, result.ReachErr, reset)
	case !result.Authenticated:
		fmt.Printf("%s⚠%s  joignable — authentification échouée : %s%v%s\n", yellow, reset, dim, result.AuthErr, reset)
		hint(importHint(ctxName))
	default:
		ok("joignable et authentifié")
	}
	return nil
}
