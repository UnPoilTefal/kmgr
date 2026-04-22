package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var renameCmd = &cobra.Command{
	Use:               "rename <old-context> <new-context>",
	Short:             "Renomme un contexte kubeconfig (ex: john@staging → john@prod)",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeContexts,
	RunE:              runRename,
}

func runRename(_ *cobra.Command, args []string) error {
	oldCtx := args[0]
	newCtx := args[1]

	section(fmt.Sprintf("Renommage : %s → %s", oldCtx, newCtx))

	// Validate old context format.
	if !strings.Contains(oldCtx, "@") {
		return fmt.Errorf("format attendu : <user>@<cluster>, reçu : %s", oldCtx)
	}

	// Validate new context format.
	at := strings.Index(newCtx, "@")
	if at < 0 {
		return fmt.Errorf("format attendu : <user>@<cluster>, reçu : %s", newCtx)
	}

	_, configsDir, _ := config.Dirs()
	oldFile := filepath.Join(configsDir, fmt.Sprintf("kubeconfig_%s.yaml", oldCtx))
	newFile := filepath.Join(configsDir, fmt.Sprintf("kubeconfig_%s.yaml", newCtx))

	// Source must exist.
	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		return fmt.Errorf("fichier non trouvé : %s", oldFile)
	}

	// Destination must not exist.
	if _, err := os.Stat(newFile); err == nil {
		return fmt.Errorf("le fichier de destination existe déjà : %s", newFile)
	}

	newCluster := newCtx[at+1:]
	// AuthInfo is namespaced as the full context name.
	if _, _, _, err := config.NormalizeAndWrite(oldFile, newFile, newCtx, newCluster, newCtx); err != nil {
		return fmt.Errorf("erreur lors de la normalisation : %w", err)
	}

	if err := os.Remove(oldFile); err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'ancien fichier : %w", err)
	}
	ok(fmt.Sprintf("Renommé : %s → %s", oldCtx, newCtx))

	return runMergeInternal()
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
