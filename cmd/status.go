package cmd

import (
	"fmt"

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
	result := config.StatusResult{
		ContextName:  ctxName,
		ClusterName:  clusterName,
		Server:       server,
		Connectivity: config.TestConnectivity(),
	}
	fmt.Print(result.Render(outputMode(), palette()))
	if !aiMode && result.Connectivity.Reachable && !result.Connectivity.Authenticated {
		hint(importHint(ctxName))
	}
	return nil
}
