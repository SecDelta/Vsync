package cmd

import (
	"log"
	"os"

	"github.com/SecDelta/Vsync/meta"
	kv "github.com/SecDelta/Vsync/pkg"
	"github.com/spf13/cobra"
)

var (
	srcVault, destVault, srcToken, destToken, kvPath string
)

var rootCmd = &cobra.Command{
	Use:     "Vsync",
	Short:   "Vault Sync is a CLI tool to replicate secrets between two Vault instances.",
	Long:    `A Fast and Flexible Vault secrets replicator built with love by Go. It helps in implementing DR for Vault.`,
	Version: meta.GetVersion(),
}

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Replicate KV secrets from one Vault to another",
	Long:  `Replicate all KV secrets from a source Vault instance to a destination Vault instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if srcToken == "" {
			srcToken = os.Getenv("SRC_VAULT_TOKEN")
		}

		if destToken == "" {
			destToken = os.Getenv("DEST_VAULT_TOKEN")
		}

		if srcToken == "" || destToken == "" {
			log.Fatalf("Vault tokens are required but not provided")
		}

		if kvPath == "" {
			kvPath = "secret"
		}

		if err := kv.ReplicateKVSecrets(srcVault, destVault, srcToken, destToken, kvPath); err != nil {
			log.Fatalf("Error replicating KV secrets: %s", err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(kvCmd)
	kvCmd.Flags().StringVarP(&srcVault, "src-vault", "s", "", "Source Vault address (required)")
	kvCmd.Flags().StringVarP(&destVault, "dest-vault", "d", "", "Destination Vault address (required)")
	kvCmd.Flags().StringVarP(&srcToken, "src-token", "", "", "Source Vault token")
	kvCmd.Flags().StringVarP(&destToken, "dest-token", "", "", "Destination Vault token")
	kvCmd.Flags().StringVarP(&kvPath, "path", "p", "secret", "KV engine path (e.g., 'secret')")
	kvCmd.MarkFlagRequired("src-vault")
	kvCmd.MarkFlagRequired("dest-vault")
}
