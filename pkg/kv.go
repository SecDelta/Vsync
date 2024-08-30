package kv

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/vault/api"
)

// ReplicateKVSecrets connects to both Vault instances and replicates the KV secrets.
func ReplicateKVSecrets(srcAddr, destAddr, srcToken, destToken, kvPath string) error {
	srcClient, err := createVaultClient(srcAddr, srcToken)
	if err != nil {
		return fmt.Errorf("error creating source Vault client: %v", err)
	}

	destClient, err := createVaultClient(destAddr, destToken)
	if err != nil {
		return fmt.Errorf("error creating destination Vault client: %v", err)
	}

	kvPath = strings.TrimSuffix(kvPath, "/") // Ensure no trailing slash

	log.Printf("Starting replication from %s to %s under KV path '%s'", srcAddr, destAddr, kvPath)

	// Start the replication process from the root of the specified KV engine
	if err := replicatePath(srcClient, destClient, kvPath+"/metadata/", kvPath+"/data/"); err != nil {
		return fmt.Errorf("error during replication: %v", err)
	}

	log.Println("Replication completed successfully.")
	return nil
}

// createVaultClient creates and configures a Vault client.
func createVaultClient(address, token string) (*api.Client, error) {
	client, err := api.NewClient(&api.Config{Address: address})
	if err != nil {
		return nil, fmt.Errorf("error creating Vault client for address %s: %v", address, err)
	}
	client.SetToken(token)
	log.Printf("Vault client created for address %s", address)
	return client, nil
}

// listSecrets lists the secrets at the given metadata path.
func listSecrets(client *api.Client, path string) ([]string, error) {
	log.Printf("Listing secrets at path: %s", path)
	secret, err := client.Logical().List(path)
	if err != nil {
		log.Printf("Error listing secrets at path %s: %v", path, err)
		return nil, err
	}
	if secret == nil || secret.Data["keys"] == nil {
		log.Printf("No keys found at path %s", path)
		return []string{}, nil
	}

	var results []string
	for _, k := range secret.Data["keys"].([]interface{}) {
		results = append(results, k.(string))
	}
	log.Printf("Found keys: %v at path: %s", results, path)
	return results, nil
}

// readSecret reads the secret at the given data path.
func readSecret(client *api.Client, path string) (map[string]interface{}, error) {
	dataPath := strings.Replace(path, "metadata", "data", 1)
	log.Printf("Reading secret at data path: %s", dataPath)
	secret, err := client.Logical().Read(dataPath)
	if err != nil {
		log.Printf("Error reading secret at path %s: %v", dataPath, err)
		return nil, err
	}
	if secret == nil || secret.Data["data"] == nil {
		log.Printf("No data found at path %s", dataPath)
		return nil, fmt.Errorf("no data found at path: %s", dataPath)
	}
	log.Printf("Read secret data from path: %s", dataPath)
	return secret.Data["data"].(map[string]interface{}), nil
}

// writeSecret writes the secret data to the given data path.
func writeSecret(client *api.Client, path string, data map[string]interface{}) error {
	log.Printf("Writing secret to path: %s", path)
	_, err := client.Logical().Write(path, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		log.Printf("Error writing secret to path %s: %v", path, err)
		return err
	}
	log.Printf("Successfully wrote secret to path: %s", path)
	return nil
}

// replicatePath recursively replicates the secrets from the source to the destination Vault.
func replicatePath(srcClient, destClient *api.Client, metadataPath, dataPath string) error {
	log.Printf("Replicating from metadata path: %s to data path: %s", metadataPath, dataPath)
	secrets, err := listSecrets(srcClient, metadataPath)
	if err != nil {
		log.Printf("Error listing secrets at %s: %v", metadataPath, err)
		return err
	}

	for _, secretKey := range secrets {
		log.Printf("Processing secret key: %s", secretKey)
		if strings.HasSuffix(secretKey, "/") {
			newMetadataPath := metadataPath + secretKey
			newDataPath := dataPath + secretKey
			log.Printf("Recursing into path: %s", newMetadataPath)
			err := replicatePath(srcClient, destClient, newMetadataPath, newDataPath)
			if err != nil {
				return fmt.Errorf("error replicating subpath %s: %v", newMetadataPath, err)
			}
		} else {
			fullSrcPath := metadataPath + secretKey
			fullDestPath := dataPath + secretKey

			log.Printf("Reading secret from source path: %s", fullSrcPath)
			secretData, err := readSecret(srcClient, fullSrcPath)
			if err != nil {
				return fmt.Errorf("error reading secret at %s: %v", fullSrcPath, err)
			}

			log.Printf("Writing secret to destination path: %s", fullDestPath)
			err = writeSecret(destClient, fullDestPath, secretData)
			if err != nil {
				return fmt.Errorf("error writing secret to %s: %v", fullDestPath, err)
			}

			log.Printf("Successfully replicated secret from %s to %s", fullSrcPath, fullDestPath)
		}
	}
	return nil
}
