package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
	"github.com/UnPoilTefal/kmgr/internal/normalize"
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
	if !aiMode {
		fmt.Printf("  %s%s%s\n\n", dim, configsDir, reset)
	}

	sources, err := config.CheckSourceFiles(configsDir)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		warn("Aucun kubeconfig_*.yaml — lance : kmgr import -f <fichier> -u <user> -c <cluster>")
	}
	sourceIssues := 0
	for _, s := range sources {
		printSourceCheck(s)
		sourceIssues += len(s.Issues)
	}

	// ---- Section 2 : kubeconfig cible ----------------------------------------
	section("Kubeconfig cible")
	if !aiMode {
		fmt.Printf("  %s%s%s\n\n", dim, mergedPath, reset)
	}

	target, err := config.CheckTarget(mergedPath)
	if err != nil {
		return err
	}
	printTargetCheck(target)

	// ---- Résumé ---------------------------------------------------------------
	if !aiMode {
		fmt.Println()
	}
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
		hint("kmgr fix")
	}
	if contextIssues > 0 {
		warn(fmt.Sprintf("%d problème(s) structurel(s) dans le fichier cible", contextIssues))
		hint("kmgr merge")
	}
	if unreachable > 0 {
		warn(fmt.Sprintf("%d contexte(s) non joignable(s) — vérifier la connectivité réseau", unreachable))
	}
	if unauthenticated > 0 {
		warn(fmt.Sprintf("%d contexte(s) joignable(s) mais authentification échouée", unauthenticated))
		hint("kmgr import --force -u <user> -c <cluster>")
	}

	// Exit code 1 pour permettre l'usage en CI / scripting.
	os.Exit(1)
	return nil
}

// printSourceCheck affiche le résultat d'un fichier source.
func printSourceCheck(s config.SourceCheck) {
	mode := outputMode()
	fmt.Print(s.Render(mode, palette()))
	if mode == config.AI || s.OK() {
		return
	}
	if s.UnfixableReason != "" {
		hint(fmt.Sprintf("kmgr fix (%s → mise en quarantaine)", s.UnfixableReason))
	} else {
		hint("kmgr fix")
	}
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
	mode := outputMode()
	fmt.Print(c.Render(mode, palette()))
	if mode == config.AI {
		return
	}
	if len(c.Issues) > 0 {
		hint("kmgr merge")
	}
	if c.Reachable && !c.Authenticated {
		hint(importHint(c.ContextName))
	}
	fmt.Println()
}

// importHint retourne la commande import --force avec user et cluster dérivés du contexte.
func importHint(ctxName string) string {
	identity, ok := normalize.Parse(ctxName)
	if !ok {
		return "kmgr import --force -u <user> -c <cluster>"
	}
	return fmt.Sprintf("kmgr import --force -u %s -c %s", identity.User(), identity.Cluster())
}
