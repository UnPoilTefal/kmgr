package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Re-merge tous les kubeconfigs sources",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runMergeInternal()
	},
}

func runMergeInternal() error {
	section("Merge des kubeconfigs")

	bak, err := config.BackupMerged()
	if err != nil {
		return err
	}
	if bak != "" {
		info(fmt.Sprintf("Backup créé : %s", bak))
	}

	_, configsDir, _ := config.Dirs()
	result, err := config.MergeAll(configsDir, config.MergedFile())
	if err != nil {
		return err
	}
	if result == nil {
		warn(fmt.Sprintf("Aucun fichier kubeconfig_*.yaml trouvé dans %s", configsDir))
		return nil
	}

	ok(fmt.Sprintf("%d fichier(s) mergé(s) → %s", len(result.Files), config.MergedFile()))
	for _, f := range result.Files {
		fmt.Printf("  %s+%s %s\n", dim, reset, filepath.Base(f))
	}
	if len(result.Quarantined) > 0 {
		_, configsDir, _ := config.Dirs()
		quarantineDir := filepath.Join(configsDir, "quarantine")
		for _, f := range result.Quarantined {
			logErr(fmt.Sprintf("%s : nom non conforme ou non parseable → mis en quarantaine", f))
			hint(fmt.Sprintf("kcfg import -f %s -u <user> -c <cluster>", filepath.Join(quarantineDir, f)))
		}
	}

	if result.RestoredCtx != "" {
		switch {
		case result.RestoredOnline && result.RestoredAuthed:
			ok(fmt.Sprintf("Contexte actif restauré : %s (joignable et authentifié)", result.RestoredCtx))
		case result.RestoredOnline:
			warn(fmt.Sprintf("Contexte actif restauré : %s (joignable, authentification échouée)", result.RestoredCtx))
		default:
			warn(fmt.Sprintf("Contexte actif restauré : %s (non joignable)", result.RestoredCtx))
		}
	}
	return nil
}
