package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
	"github.com/UnPoilTefal/kmgr/internal/normalize"
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

	oldIdentity, valid := normalize.Parse(oldCtx)
	if !valid {
		return fmt.Errorf("format attendu : <user>@<cluster>, reçu : %s", oldCtx)
	}
	newIdentity, valid := normalize.Parse(newCtx)
	if !valid {
		return fmt.Errorf("format attendu : <user>@<cluster>, reçu : %s", newCtx)
	}

	_, configsDir, _ := config.Dirs()
	oldFile := filepath.Join(configsDir, oldIdentity.SourceFilename())
	newFile := filepath.Join(configsDir, newIdentity.SourceFilename())

	// Source must exist.
	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		return fmt.Errorf("fichier non trouvé : %s", oldFile)
	}

	// Destination must not exist.
	if _, err := os.Stat(newFile); err == nil {
		return fmt.Errorf("le fichier de destination existe déjà : %s", newFile)
	}

	if _, _, _, err := config.NormalizeAndWrite(oldFile, newFile, newIdentity, newIdentity.String()); err != nil {
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
