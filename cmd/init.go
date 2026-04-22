package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise la structure de répertoires",
	RunE:  runInit,
}

func runInit(_ *cobra.Command, _ []string) error {
	section("Initialisation de kcfg")

	_, configsDir, backupDir := config.Dirs()

	for _, dir := range []string{configsDir, backupDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	// Add KUBECONFIG export and shell completion to the user's shell profile if absent.
	profile, shellName := shellInfo()
	data, _ := os.ReadFile(profile)
	content := string(data)

	kcLine := `export KUBECONFIG="${KCFG_DIR:-$HOME/.kube}/config"`
	completionLine := fmt.Sprintf("source <(kcfg completion %s)", shellName)

	if strings.Contains(content, "KUBECONFIG") {
		info(fmt.Sprintf("KUBECONFIG déjà présent dans %s", profile))
	} else {
		if f, err := os.OpenFile(profile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			warn(fmt.Sprintf("Impossible d'ouvrir %s : %v", profile, err))
		} else {
			fmt.Fprintf(f, "\n# kcfg\n%s\n", kcLine)
			f.Close()
			ok(fmt.Sprintf("KUBECONFIG ajouté à %s (recharge ton shell)", profile))
		}
	}

	if strings.Contains(content, "kcfg completion") {
		info(fmt.Sprintf("Complétion déjà configurée dans %s", profile))
	} else {
		if f, err := os.OpenFile(profile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			warn(fmt.Sprintf("Impossible d'ajouter la complétion à %s : %v", profile, err))
		} else {
			fmt.Fprintf(f, "%s\n", completionLine)
			f.Close()
			ok(fmt.Sprintf("Complétion %s ajoutée à %s (recharge ton shell)", shellName, profile))
		}
	}

	kubeDir, _, _ := config.Dirs()
	ok(fmt.Sprintf("Structure prête : %s", kubeDir))
	fmt.Printf("  %sconfigs/%s  → fichiers sources individuels\n", cyan, reset)
	fmt.Printf("  %sconfig%s    → fichier mergé actif\n", cyan, reset)
	fmt.Printf("  %sbackups/%s  → sauvegardes automatiques\n", cyan, reset)
	return nil
}

func shellInfo() (profile, name string) {
	home, _ := os.UserHomeDir()
	if strings.HasSuffix(os.Getenv("SHELL"), "zsh") {
		return filepath.Join(home, ".zshrc"), "zsh"
	}
	return filepath.Join(home, ".bashrc"), "bash"
}
