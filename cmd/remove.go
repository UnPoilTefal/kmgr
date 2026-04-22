package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnPoilTefal/kmgr/internal/config"
)

var removeCmd = &cobra.Command{
	Use:               "remove <contexte>",
	Short:             "Supprime un kubeconfig (ex: john@prod-payments)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeContexts,
	RunE:              runRemove,
}

func runRemove(_ *cobra.Command, args []string) error {
	ctxName := args[0]
	section(fmt.Sprintf("Suppression : %s", ctxName))

	// Derive user and cluster from "user@cluster".
	at := strings.Index(ctxName, "@")
	if at < 0 {
		return fmt.Errorf("format attendu : <user>@<cluster>, reçu : %s", ctxName)
	}
	user := ctxName[:at]
	cluster := ctxName[at+1:]

	_, configsDir, _ := config.Dirs()
	targetFile := filepath.Join(configsDir, fmt.Sprintf("kubeconfig_%s@%s.yaml", user, cluster))

	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return fmt.Errorf("fichier non trouvé : %s", targetFile)
	}

	fmt.Printf("Supprimer %s ? [y/N] ", targetFile)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" {
		info("Annulé.")
		return nil
	}

	if err := os.Remove(targetFile); err != nil {
		return err
	}
	ok(fmt.Sprintf("Supprimé : %s", targetFile))
	return runMergeInternal()
}
