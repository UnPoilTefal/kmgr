package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

// ---------------------------------------------------------------------------
// Palette ANSI — respecte NO_COLOR (https://no-color.org)
// ---------------------------------------------------------------------------

var (
	reset  = "\033[0m"
	red    = "\033[0;31m"
	yellow = "\033[1;33m"
	green  = "\033[0;32m"
	cyan   = "\033[0;36m"
	dim    = "\033[2m"
	bold   = "\033[1m"
)

func initColors() {
	if os.Getenv("NO_COLOR") != "" || quietMode {
		reset = ""
		red = ""
		yellow = ""
		green = ""
		cyan = ""
		dim = ""
		bold = ""
	}

}

// ---------------------------------------------------------------------------
// Helpers d'affichage
// ---------------------------------------------------------------------------

var quietMode bool

// stdout — silencieux en mode quiet.
func info(msg string)    { ifVerbose(fmt.Sprintf("%s▸%s %s\n", cyan, reset, msg)) }
func ok(msg string)      { ifVerbose(fmt.Sprintf("%s✓%s %s\n", green, reset, msg)) }
func warn(msg string)    { ifVerbose(fmt.Sprintf("%s⚠%s %s\n", yellow, reset, msg)) }
func section(msg string) { ifVerbose(fmt.Sprintf("\n%s%s%s\n", bold, msg, reset)) }
func hint(cmd string)    { ifVerbose(fmt.Sprintf("  %s→%s  %s\n", cyan, reset, cmd)) }

// stderr — toujours affiché, même en mode quiet.
func logErr(msg string) { fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, msg) }

func ifVerbose(s string) {
	if !quietMode {
		fmt.Print(s)
	}
}

// warnEnvDesync affiche un avertissement si KUBECONFIG ne pointe pas vers
// le fichier mergé géré par kcfg (désynchro entre env et KCFG_DIR).
func warnEnvDesync() {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		return
	}
	merged := config.MergedFile()
	for _, p := range filepath.SplitList(kc) {
		if p == merged {
			return // cohérent
		}
	}
	warn(fmt.Sprintf("KUBECONFIG=%s ne pointe pas vers le fichier kcfg (%s)", kc, merged))
	hint(fmt.Sprintf("export KUBECONFIG=%s", merged))
}

// ---------------------------------------------------------------------------
// Root command
// ---------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Use:   "kcfg",
	Short: "Kubeconfig Manager — normalise et gère les kubeconfigs d'entreprise",
	Long: `kcfg — Kubeconfig Manager

Convention de nommage :
  Fichier source : kubeconfig_{user}@{cluster}.yaml
  Contexte       : {user}@{cluster}
  Exemple        : kubeconfig_john@prod-payments.yaml → john@prod-payments

Variables d'environnement :
  KCFG_DIR   Répertoire de base (défaut: ~/.kube)
  NO_COLOR   Désactive les couleurs (https://no-color.org)`,
	SilenceErrors: true, // on gère l'affichage nous-mêmes
	SilenceUsage:  true, // on affiche le usage explicitement si besoin
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		initColors()
	},
}

// lastCmd retient la commande effectivement exécutée pour pouvoir
// afficher son usage en cas d'erreur (PersistentPreRun est appelé avant RunE).
var lastCmd *cobra.Command

func Execute() {
	initColors() // nécessaire pour les erreurs avant PersistentPreRun (ex: flags manquants)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		lastCmd = cmd
		initColors()
	}

	if err := rootCmd.Execute(); err != nil {
		if isUsageError(err) {
			target := lastCmd
			if target == nil {
				target = rootCmd
			}
			_ = target.Usage()
			fmt.Fprintln(os.Stderr)
		}
		logErr(err.Error())
		os.Exit(1)
	}
}

// isUsageError détecte les erreurs liées à un mauvais usage de la CLI
// (flag manquant, argument inconnu, valeur invalide…).
func isUsageError(err error) bool {
	msg := err.Error()
	for _, prefix := range []string{
		"required flag",
		"unknown flag",
		"unknown command",
		"invalid argument",
		"accepts ",
	} {
		if len(msg) >= len(prefix) && msg[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Supprime toute sortie sauf les erreurs")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(statusCmd)
}
