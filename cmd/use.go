package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var useCmd = &cobra.Command{
	Use:               "use <contexte>",
	Short:             "Switche le contexte actif",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeContexts,
	RunE:              runUse,
}

func runUse(_ *cobra.Command, args []string) error {
	ctxName := args[0]
	if err := config.SetCurrentContext(ctxName); err != nil {
		return err
	}
	ok(fmt.Sprintf("Contexte actif → %s", ctxName))
	return nil
}
