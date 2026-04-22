package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
	"github.com/UnPoilTefal/kmgr/internal/normalize"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Importe et normalise un kubeconfig",
	Long: `Importe un fichier kubeconfig, normalise les noms de contexte/cluster/user
selon la convention {user}@{cluster}, et le stocke dans ~/.kube/configs/.

La source peut être un fichier (-f), le contenu du clipboard (--clipboard)
ou un pipe stdin (--stdin).

Exemples :
  kcfg import -f ~/Downloads/kubeconfig.yaml -u john -c prod-payments
  kcfg import --clipboard -u john -c prod-payments
  k3d kubeconfig get mycluster  | kcfg import --stdin -u john -c mycluster
  kind get kubeconfig --name dev | kcfg import --stdin -u john -c dev`,
	RunE: runImport,
}

var (
	importFile      string
	importClipboard bool
	importStdin     bool
	importUser      string
	importCluster   string
	importCtx       string
	importForce     bool
)

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "Fichier kubeconfig source")
	importCmd.Flags().BoolVarP(&importClipboard, "clipboard", "p", false, "Lire le kubeconfig depuis le clipboard")
	importCmd.Flags().BoolVar(&importStdin, "stdin", false, "Lire le kubeconfig depuis stdin (pipe)")
	importCmd.Flags().StringVarP(&importUser, "user", "u", "", "Nom d'utilisateur (requis)")
	importCmd.Flags().StringVarP(&importCluster, "cluster", "c", "", "Nom du cluster (requis)")
	importCmd.Flags().StringVar(&importCtx, "ctx", "", "Nom de contexte personnalisé (défaut: {user}@{cluster})")
	importCmd.Flags().BoolVar(&importForce, "force", false, "Écraser si déjà importé (backup automatique)")
	_ = importCmd.MarkFlagRequired("user")
	_ = importCmd.MarkFlagRequired("cluster")
}

func runImport(_ *cobra.Command, _ []string) error {
	// Validate source flags — exactly one source must be specified.
	sources := 0
	if importFile != "" {
		sources++
	}
	if importClipboard {
		sources++
	}
	if importStdin {
		sources++
	}
	if sources > 1 {
		return fmt.Errorf("--file, --clipboard et --stdin sont mutuellement exclusifs")
	}
	if sources == 0 {
		return fmt.Errorf("une source est requise : --file <fichier>, --clipboard ou --stdin")
	}

	// Resolve the source to a temp file.
	srcFile := importFile
	if importClipboard {
		tmp, err := writeClipboardToTemp()
		if err != nil {
			return err
		}
		defer os.Remove(tmp)
		srcFile = tmp
		info("Contenu du clipboard écrit dans un fichier temporaire")
	} else if importStdin {
		tmp, err := writeStdinToTemp()
		if err != nil {
			return err
		}
		defer os.Remove(tmp)
		srcFile = tmp
		info("Contenu stdin écrit dans un fichier temporaire")
	} else {
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			return fmt.Errorf("fichier introuvable : %s", srcFile)
		}
	}

	// Validate kubeconfig before doing anything else.
	if err := config.ValidateKubeconfig(srcFile); err != nil {
		return fmt.Errorf("kubeconfig invalide : %w", err)
	}
	ok("Fichier kubeconfig valide")

	user := normalize.Name(importUser)
	cluster := normalize.Name(importCluster)
	ctxName := importCtx
	if ctxName == "" {
		ctxName = user + "@" + cluster
	}
	// AuthInfo est namespaced par cluster pour éviter les collisions au merge.
	authInfo := user + "@" + cluster

	_, configsDir, _ := config.Dirs()
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		return err
	}

	destFile := filepath.Join(configsDir, fmt.Sprintf("kubeconfig_%s@%s.yaml", user, cluster))

	if _, err := os.Stat(destFile); err == nil {
		if !importForce {
			return fmt.Errorf("fichier déjà importé : %s\nUtilise --force pour écraser", destFile)
		}
		bak, err := config.BackupFile(destFile)
		if err != nil {
			return fmt.Errorf("backup impossible : %w", err)
		}
		info(fmt.Sprintf("Backup créé : %s", bak))
	}

	section(fmt.Sprintf("Import : %s", ctxName))
	info("Normalisation des noms (contexte, cluster, user)...")

	oldCtx, oldCluster, oldUser, err := config.NormalizeAndWrite(srcFile, destFile, ctxName, cluster, authInfo)
	if err != nil {
		return err
	}

	fmt.Printf("  %sancien contexte%s : %s\n", dim, reset, oldCtx)
	fmt.Printf("  %sancien cluster %s : %s\n", dim, reset, oldCluster)
	fmt.Printf("  %sancien user    %s : %s\n", dim, reset, oldUser)
	fmt.Printf("  %s→%s contexte      : %s%s%s\n", green, reset, bold, ctxName, reset)
	fmt.Printf("  %s→%s cluster       : %s\n", green, reset, cluster)
	fmt.Printf("  %s→%s user          : %s\n", green, reset, authInfo)
	fmt.Printf("  %s→%s fichier       : %s%s%s\n", green, reset, dim, destFile, reset)

	ok("Import terminé")
	return runMergeInternal()
}

// writeStdinToTemp reads stdin and writes its content to a temp file.
// The caller is responsible for removing the file.
func writeStdinToTemp() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("lecture stdin : %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("stdin est vide")
	}
	tmp, err := os.CreateTemp("", "kcfg-stdin-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}

// writeClipboardToTemp reads the clipboard and writes its content to a temp file.
// The caller is responsible for removing the file.
func writeClipboardToTemp() (string, error) {
	content, err := config.ReadClipboard()
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "kcfg-clipboard-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}
