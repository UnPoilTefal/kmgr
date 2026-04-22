package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Corrige les fichiers sources dont les noms internes ne correspondent pas au nom de fichier",
	Long: `Pour chaque kubeconfig_*.yaml dans ~/.kube/configs/ :
  • renomme le contexte, le cluster et l'user pour qu'ils correspondent au nom de fichier
  • corrige les permissions à 0600

Aucun fichier n'est modifié s'il est déjà conforme.
Un merge est lancé automatiquement si au moins un fichier a été corrigé.`,
	RunE: runFix,
}

func init() {
	rootCmd.AddCommand(fixCmd)
}

func runFix(_ *cobra.Command, _ []string) error {
	section("Correction des fichiers source")

	_, configsDir, _ := config.Dirs()

	sources, err := config.CheckSourceFiles(configsDir)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		warn("Aucun kubeconfig_*.yaml trouvé.")
		return nil
	}

	fixed := 0
	errors := 0
	quarantined := 0
	for _, s := range sources {
		fullPath := filepath.Join(configsDir, s.File)

		// Un fichier non corrigeable automatiquement est mis en quarantaine :
		//   • nom sans user@cluster (ex: kubeconfig_monocluster.yaml)
		//   • contenu non parseable
		unfixable := !normalize.IsValidSourceFilename(fullPath)
		if !unfixable {
			for _, issue := range s.Issues {
				if issue.Field == "parse" {
					unfixable = true
					break
				}
			}
		}

		if unfixable {
			dest, err := config.QuarantineFile(fullPath)
			if err != nil {
				logErr(fmt.Sprintf("%s : quarantaine impossible : %v", s.File, err))
				errors++
			} else {
				warn(fmt.Sprintf("%s : non corrigeable → mis en quarantaine", s.File))
				hint(fmt.Sprintf("kcfg import -f %s -u <user> -c <cluster>", dest))
				quarantined++
			}
			continue
		}

		changed, err := fixSourceFile(configsDir, s)
		if err != nil {
			logErr(fmt.Sprintf("%s : %v", s.File, err))
			errors++
			continue
		}
		if changed {
			fixed++
		}
	}

	fmt.Println()
	if quarantined > 0 {
		warn(fmt.Sprintf("%d fichier(s) mis en quarantaine", quarantined))
	}
	if errors > 0 {
		warn(fmt.Sprintf("%d fichier(s) non corrigeable(s) — correction manuelle requise", errors))
	}
	if fixed == 0 && errors == 0 && quarantined == 0 {
		ok("Tous les fichiers sont déjà conformes.")
		return nil
	}
	if fixed > 0 || quarantined > 0 {
		if fixed > 0 {
			ok(fmt.Sprintf("%d fichier(s) corrigé(s)", fixed))
		}
		return runMergeInternal()
	}
	return nil
}

// fixSourceFile corrige un fichier source si nécessaire.
// Retourne true si le fichier a été modifié.
func fixSourceFile(configsDir string, s config.SourceCheck) (bool, error) {
	if s.OK() {
		ok(s.File)
		return false, nil
	}

	// Noms attendus d'après le nom de fichier.
	fullPath := configsDir + "/" + s.File
	expectedCtx := normalize.ContextFromFile(fullPath)
	at := strings.LastIndex(expectedCtx, "@")
	if at < 0 {
		return false, fmt.Errorf("impossible de dériver user@cluster depuis %q", s.File)
	}
	// AuthInfo est namespaced : identique au nom du contexte (ex: deschampsf@ctain-d-00).
	expectedCluster := expectedCtx[at+1:]
	expectedUser := expectedCtx // AuthInfo == ctxName

	// Un fichier non parseable ne peut pas être corrigé automatiquement.
	for _, issue := range s.Issues {
		if issue.Field == "parse" {
			warn(fmt.Sprintf("%s : non parseable, correction manuelle requise (%s)", s.File, issue.Got))
			return false, nil
		}
	}

	oldCtx, oldCluster, oldUser, err := config.NormalizeAndWrite(fullPath, fullPath, expectedCtx, expectedCluster, expectedUser)
	if err != nil {
		return false, err
	}

	// Affiche uniquement les champs qui ont réellement changé.
	fmt.Printf("  %s~%s %s%s%s\n", yellow, reset, bold, s.File, reset)
	printIfChanged("context", oldCtx, expectedCtx)
	printIfChanged("cluster", oldCluster, expectedCluster)
	printIfChanged("user", oldUser, expectedUser)

	// Signale si les permissions ont aussi été corrigées.
	for _, issue := range s.Issues {
		if issue.Field == "permissions" {
			fi, _ := os.Stat(fullPath)
			if fi != nil && fi.Mode().Perm() == 0600 {
				fmt.Printf("      permissions : %s%s → 0600%s\n", yellow, issue.Got, reset)
			}
		}
	}

	return true, nil
}

func printIfChanged(field, old, new string) {
	if old != new {
		fmt.Printf("      %s : %s%s%s → %s%s%s\n",
			field,
			yellow, old, reset,
			green, new, reset,
		)
	}
}
