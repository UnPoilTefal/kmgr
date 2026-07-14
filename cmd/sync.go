package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Gère les copies miroir du kubeconfig mergé (ex: côté Windows depuis WSL2)",
	Long: `Gère des copies miroir du kubeconfig mergé vers des chemins externes.

Chaque miroir est resynchronisé automatiquement après toute écriture du
fichier mergé (import, merge, use, rename, remove, fix). kmgr étant le seul
écrivain de ce fichier, la copie est one-way : ne modifie pas les miroirs
à la main.

Cas d'usage WSL2 : publier le kubeconfig côté Windows pour que les outils
Windows (Lens, kubectl.exe…) voient toujours la dernière version :

  kmgr sync add --windows       # auto-détecte /mnt/c/Users/<user>/.kube/config
  kmgr sync add /mnt/c/Users/john/.kube/config

Les cibles sont stockées dans $KMGR_DIR/mirrors (une par ligne).`,
	RunE: runSyncStatus,
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Affiche les miroirs configurés et leur état",
	Args:  cobra.NoArgs,
	RunE:  runSyncStatus,
}

var syncAddWindows bool

var syncAddCmd = &cobra.Command{
	Use:   "add [chemin]",
	Short: "Ajoute une cible miroir (--windows pour auto-détecter le profil Windows sous WSL2)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSyncAdd,
}

var syncRemoveCmd = &cobra.Command{
	Use:   "remove <chemin>",
	Short: "Retire une cible miroir (le fichier cible n'est pas supprimé)",
	Args:  cobra.ExactArgs(1),
	RunE:  runSyncRemove,
}

var syncNowCmd = &cobra.Command{
	Use:   "now",
	Short: "Resynchronise immédiatement tous les miroirs",
	Args:  cobra.NoArgs,
	RunE:  runSyncNow,
}

func init() {
	syncAddCmd.Flags().BoolVar(&syncAddWindows, "windows", false, "Auto-détecte %USERPROFILE%\\.kube\\config (WSL2 uniquement)")
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncAddCmd)
	syncCmd.AddCommand(syncRemoveCmd)
	syncCmd.AddCommand(syncNowCmd)
	rootCmd.AddCommand(syncCmd)
}

func runSyncStatus(_ *cobra.Command, _ []string) error {
	targets, err := config.ListMirrors()
	if err != nil {
		return err
	}

	section("Miroirs du kubeconfig mergé")
	if len(targets) == 0 {
		info("Aucun miroir configuré.")
		hint("kmgr sync add <chemin>   # ou : kmgr sync add --windows (WSL2)")
		return nil
	}

	stale := 0
	for _, t := range targets {
		state, err := config.CheckMirror(t)
		if aiMode {
			fmt.Printf("mirror: %s %s\n", t, state)
			if state != config.MirrorInSync {
				stale++
			}
			continue
		}
		switch state {
		case config.MirrorInSync:
			fmt.Printf("  %s✓%s %s %s(à jour)%s\n", green, reset, t, dim, reset)
		case config.MirrorStale:
			fmt.Printf("  %s⚠%s %s %s(obsolète)%s\n", yellow, reset, t, dim, reset)
			stale++
		case config.MirrorMissing:
			fmt.Printf("  %s⚠%s %s %s(absent)%s\n", yellow, reset, t, dim, reset)
			stale++
		default:
			fmt.Printf("  %s✗%s %s %s(%v)%s\n", red, reset, t, dim, err, reset)
			stale++
		}
	}
	if stale > 0 {
		hint("kmgr sync now")
	}
	return nil
}

func runSyncAdd(_ *cobra.Command, args []string) error {
	var target string
	switch {
	case syncAddWindows && len(args) > 0:
		return fmt.Errorf("--windows et <chemin> sont mutuellement exclusifs")
	case syncAddWindows:
		detected, err := config.DetectWindowsKubeconfig()
		if err != nil {
			return err
		}
		target = detected
		info(fmt.Sprintf("Profil Windows détecté : %s", target))
	case len(args) == 1:
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		target = abs
	default:
		return fmt.Errorf("une cible est requise : kmgr sync add <chemin> ou kmgr sync add --windows")
	}

	if err := config.AddMirror(target); err != nil {
		return err
	}
	ok(fmt.Sprintf("Miroir ajouté : %s", target))

	// Synchronisation immédiate pour que la cible soit à jour dès maintenant.
	reportSyncResults(config.SyncMirrors())
	return nil
}

func runSyncRemove(_ *cobra.Command, args []string) error {
	abs, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	if err := config.RemoveMirror(abs); err != nil {
		return err
	}
	ok(fmt.Sprintf("Miroir retiré : %s (fichier cible conservé)", abs))
	return nil
}

func runSyncNow(_ *cobra.Command, _ []string) error {
	targets, err := config.ListMirrors()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		warn("Aucun miroir configuré.")
		hint("kmgr sync add <chemin>   # ou : kmgr sync add --windows (WSL2)")
		return nil
	}
	reportSyncResults(config.SyncMirrors())
	return nil
}

// reportSyncResults affiche le résultat d'une synchronisation des miroirs.
// Les échecs sont signalés mais jamais fatals : le fichier mergé local reste
// la référence même si une cible (ex: montage Windows) est indisponible.
func reportSyncResults(results []config.MirrorResult) {
	for _, r := range results {
		if r.Err != nil {
			warn(fmt.Sprintf("miroir non synchronisé : %s (%v)", r.Target, r.Err))
		} else {
			ok(fmt.Sprintf("Miroir synchronisé : %s", r.Target))
		}
	}
}
