package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var exportFile string

var exportCmd = &cobra.Command{
	Use:   "export <context>",
	Short: "Exporte un kubeconfig vers un fichier ou stdout",
	Long: `Exporte le kubeconfig correspondant au contexte donné.

Sans --file, écrit le contenu brut sur stdout (mode pipe) :
  kcfg export john@prod | kubectl apply --kubeconfig /dev/stdin -f manifest.yaml
  kcfg export john@prod > /tmp/kubeconfig.yaml

Avec --file, écrit vers le fichier destination avec les permissions 0600 :
  kcfg export john@prod --file ./kubeconfig-prod.yaml`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeContexts,
	RunE:              runExport,
}

func runExport(_ *cobra.Command, args []string) error {
	ctxName := args[0]

	_, configsDir, _ := config.Dirs()
	srcFile := filepath.Join(configsDir, fmt.Sprintf("kubeconfig_%s.yaml", ctxName))

	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return fmt.Errorf("fichier non trouvé : %s", srcFile)
	}

	data, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("erreur de lecture : %w", err)
	}

	if exportFile == "" {
		// Silent, pipeable output — write raw content to stdout.
		_, err = os.Stdout.Write(data)
		return err
	}

	if err := os.WriteFile(exportFile, data, 0600); err != nil {
		return fmt.Errorf("erreur d'écriture : %w", err)
	}
	ok(fmt.Sprintf("Exporté : %s → %s", ctxName, exportFile))
	return nil
}

func init() {
	exportCmd.Flags().StringVarP(&exportFile, "file", "f", "", "Fichier de destination (défaut: stdout)")
	rootCmd.AddCommand(exportCmd)
}
