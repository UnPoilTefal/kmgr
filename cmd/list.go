package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Liste les kubeconfigs gérés",
	RunE:  runList,
}

func runList(_ *cobra.Command, _ []string) error {
	section("Kubeconfigs gérés")

	contexts, err := config.ManagedContexts()
	if err != nil {
		return err
	}
	if len(contexts) == 0 {
		warn("Aucun kubeconfig importé. Lance : kmgr import -f <fichier> -u <user> -c <cluster>")
		return nil
	}

	mode := outputMode()
	p := palette()

	if mode == config.AI {
		for _, c := range contexts {
			fmt.Print(c.Render(mode, p))
		}
		return nil
	}

	fmt.Printf("  %-40s %-20s %s\n", "CONTEXTE", "CLUSTER", "FICHIER")
	fmt.Printf("  %-40s %-20s %s\n", "--------", "-------", "-------")

	for _, c := range contexts {
		fmt.Print(c.Render(mode, p))
	}

	fmt.Println()
	fmt.Printf("%sContexte actif : %s%s%s\n", dim, reset+bold, orNone(config.CurrentContext()), reset)
	return nil
}

func orNone(s string) string {
	if s == "" {
		return "<aucun>"
	}
	return s
}
