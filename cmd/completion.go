package cmd

import (
	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

// completeContexts est la ValidArgsFunction partagée par use et remove.
// Elle retourne les contextes disponibles dans le kubeconfig mergé.
func completeContexts(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return config.ListContexts(), cobra.ShellCompDirectiveNoFileComp
}
