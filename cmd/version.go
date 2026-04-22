package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

// Variables injectées à la compilation via ldflags (voir Makefile).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Affiche la version et la configuration active",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	// ---- Version ----------------------------------------------------------------
	section("Version")
	fmt.Printf("  %-16s %s%s%s\n", "binaire", bold, Version, reset)
	fmt.Printf("  %-16s %s\n", "commit", Commit)
	fmt.Printf("  %-16s %s\n", "build date", BuildDate)
	fmt.Printf("  %-16s %s/%s\n", "plateforme", runtime.GOOS, runtime.GOARCH)

	// ---- Configuration ----------------------------------------------------------
	section("Configuration")

	kcfgDir := os.Getenv("KCFG_DIR")
	if kcfgDir != "" {
		fmt.Printf("  %-16s %s%s%s %s(KCFG_DIR)%s\n", "répertoire", bold, kcfgDir, reset, dim, reset)
	} else {
		fmt.Printf("  %-16s %s %s(défaut)%s\n", "répertoire", kubeDir(), dim, reset)
	}

	kubeDir, configsDir, backupDir := config.Dirs()
	mergedFile := config.MergedFile()

	printPathStatus("config mergé", mergedFile)
	printPathStatus("sources", configsDir)
	printPathStatus("backups", backupDir)
	_ = kubeDir

	// ---- Clipboard --------------------------------------------------------------
	section("Clipboard")
	printClipboardStatus()

	return nil
}

// kubeDir retourne le répertoire de base sans passer par Dirs().
func kubeDir() string {
	_, _, _ = config.Dirs() // just to force init; we re-read below
	home, _ := os.UserHomeDir()
	return home + "/.kube"
}

// printPathStatus affiche un chemin avec un indicateur d'existence.
func printPathStatus(label, path string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  %-16s %s✓%s %s%s%s\n", label, green, reset, dim, path, reset)
	} else {
		fmt.Printf("  %-16s %s✗%s %s%s%s %s(absent)%s\n", label, yellow, reset, dim, path, reset, yellow, reset)
	}
}

// printClipboardStatus détecte et affiche l'outil clipboard disponible.
func printClipboardStatus() {
	switch runtime.GOOS {
	case "darwin":
		printTool("pbpaste", "macOS natif")
	case "windows":
		printTool("powershell.exe", "Windows natif")
	default:
		data, _ := os.ReadFile("/proc/version")
		if isWSL(string(data)) {
			printTool("powershell.exe", "WSL → Windows")
		} else {
			switch {
			case toolAvailable("wl-paste"):
				printTool("wl-paste", "Wayland")
			case toolAvailable("xclip"):
				printTool("xclip", "X11")
			case toolAvailable("xsel"):
				printTool("xsel", "X11")
			default:
				fmt.Printf("  %s✗%s aucun outil clipboard détecté\n", yellow, reset)
				hint("installe wl-paste (Wayland), xclip ou xsel (X11)")
			}
		}
	}
}

func printTool(name, context string) {
	if _, err := exec.LookPath(name); err == nil {
		fmt.Printf("  %s✓%s %-14s %s(%s)%s\n", green, reset, name, dim, context, reset)
	} else {
		fmt.Printf("  %s✗%s %-14s introuvable dans PATH\n", red, reset, name)
	}
}

func isWSL(procVersion string) bool {
	lower := strings.ToLower(procVersion)
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

func toolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
