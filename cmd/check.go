package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Vérifie l'intégrité des kubeconfigs sources et du kubeconfig cible",
	Long: `Deux vérifications distinctes :

  1. Fichiers source (configs/kubeconfig_*.yaml)
       • parsing valide
       • convention de nommage context/cluster/user vs nom de fichier
       • permissions 0600

  2. Kubeconfig cible (~/.kube/config)
       • parsing + permissions
       • cohérence structurelle de chaque contexte (cluster, user, server, credentials)
       • connectivité TCP+TLS (/version) et authentification (/api/v1) en parallèle

Retourne exit code 1 si des anomalies sont détectées.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(_ *cobra.Command, _ []string) error {
	warnEnvDesync()

	_, configsDir, _ := config.Dirs()
	mergedPath := config.MergedFile()

	// ---- Section 1 : fichiers source ----------------------------------------
	section("Fichiers source")
	fmt.Printf("  %s%s%s\n\n", dim, configsDir, reset)

	sources, err := config.CheckSourceFiles(configsDir)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		warn("Aucun kubeconfig_*.yaml — lance : kcfg import -f <fichier> -u <user> -c <cluster>")
	}
	sourceIssues := 0
	for _, s := range sources {
		printSourceCheck(s)
		sourceIssues += len(s.Issues)
	}

	// ---- Section 2 : kubeconfig cible ----------------------------------------
	section("Kubeconfig cible")
	fmt.Printf("  %s%s%s\n\n", dim, mergedPath, reset)

	target, err := config.CheckTarget(mergedPath)
	if err != nil {
		return err
	}
	printTargetCheck(target)

	// ---- Résumé ---------------------------------------------------------------
	fmt.Println()
	targetIssues := len(target.Issues)
	contextIssues := 0
	unreachable := 0
	unauthenticated := 0
	for _, c := range target.Contexts {
		targetIssues += len(c.Issues)
		contextIssues += len(c.Issues)
		if !c.Reachable {
			unreachable++
		} else if !c.Authenticated {
			unauthenticated++
		}
	}

	allOK := sourceIssues == 0 && targetIssues == 0 && unreachable == 0 && unauthenticated == 0
	if allOK {
		ok(fmt.Sprintf(
			"%d source(s), %d contexte(s) — tout est conforme, joignable et authentifié",
			len(sources), len(target.Contexts),
		))
		return nil
	}

	if sourceIssues > 0 {
		warn(fmt.Sprintf("%d problème(s) de normalisation dans les fichiers source", sourceIssues))
		hint("kcfg fix")
	}
	if contextIssues > 0 {
		warn(fmt.Sprintf("%d problème(s) structurel(s) dans le fichier cible", contextIssues))
		hint("kcfg merge")
	}
	if unreachable > 0 {
		warn(fmt.Sprintf("%d contexte(s) non joignable(s) — vérifier la connectivité réseau", unreachable))
	}
	if unauthenticated > 0 {
		warn(fmt.Sprintf("%d contexte(s) joignable(s) mais authentification échouée", unauthenticated))
		hint("kcfg import --force -u <user> -c <cluster>")
	}

	// Exit code 1 pour permettre l'usage en CI / scripting.
	os.Exit(1)
	return nil
}

// printSourceCheck affiche le résultat d'un fichier source.
func printSourceCheck(s config.SourceCheck) {
	if s.OK() {
		fmt.Printf("  %s✓%s %s\n", green, reset, s.File)
		return
	}
	fmt.Printf("  %s✗%s %s", red, reset, s.File)
	if s.Server != "" {
		fmt.Printf("  %s%s%s", dim, s.Server, reset)
	}
	fmt.Println()
	for _, issue := range s.Issues {
		fmt.Printf("      %s⚠%s  %s\n", yellow, reset, issue)
	}
	hint("kcfg fix")
	fmt.Println()
}

// printTargetCheck affiche le résultat du kubeconfig cible.
func printTargetCheck(t config.TargetCheck) {
	if len(t.Issues) > 0 {
		for _, issue := range t.Issues {
			logErr(issue.String())
		}
		return
	}

	// Trier les contextes par nom pour un affichage stable.
	sorted := make([]config.ContextCheck, len(t.Contexts))
	copy(sorted, t.Contexts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ContextName < sorted[j].ContextName
	})

	for _, c := range sorted {
		printContextCheck(c)
	}
}

// printContextCheck affiche un contexte du fichier cible.
func printContextCheck(c config.ContextCheck) {
	hasIssues := len(c.Issues) > 0

	switch {
	case !hasIssues && c.Reachable && c.Authenticated:
		fmt.Printf("  %s✓%s %s", green, reset, c.ContextName)
	case !hasIssues && c.Reachable && !c.Authenticated:
		fmt.Printf("  %s⚠%s %s", yellow, reset, c.ContextName)
	default:
		fmt.Printf("  %s✗%s %s", red, reset, c.ContextName)
	}
	if c.Server != "" {
		fmt.Printf("  %s%s%s", dim, c.Server, reset)
	}
	fmt.Println()

	for _, issue := range c.Issues {
		fmt.Printf("      %s⚠%s  %s\n", yellow, reset, issue)
	}
	if len(c.Issues) > 0 {
		hint("kcfg merge")
	}

	switch {
	case !c.Reachable:
		fmt.Printf("      %s✗%s  non joignable : %s%s%s\n", red, reset, dim, c.ReachErr, reset)
	case !c.Authenticated:
		fmt.Printf("      %s⚠%s  joignable — authentification échouée : %s%s%s\n", yellow, reset, dim, c.AuthErr, reset)
		hint(importHint(c.ContextName))
	default:
		fmt.Printf("      %s✓%s  joignable et authentifié\n", green, reset)
	}
	fmt.Println()
}

// importHint retourne la commande import --force avec user et cluster dérivés du contexte.
func importHint(ctxName string) string {
	at := strings.LastIndex(ctxName, "@")
	if at < 0 {
		return "kcfg import --force -u <user> -c <cluster>"
	}
	return fmt.Sprintf("kcfg import --force -u %s -c %s", ctxName[:at], ctxName[at+1:])
}
