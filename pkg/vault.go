package vault

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/vault/api"
)

// ReplicateKVSecrets connects to both Vault instances and replicates the KV secrets
func ReplicateKVSecrets(srcAddr, destAddr, srcToken, destToken, basePath string) error {
	srcClient, err := createVaultClient(srcAddr, srcToken)
	if err != nil {
		return fmt.Errorf("error creating source Vault client: %v", err)
	}

	destClient, err := createVaultClient(destAddr, destToken)
	if err != nil {
		return fmt.Errorf("error creating destination Vault client: %v", err)
	}

	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	log.Printf("Starting replication from %s to %s under path %s", srcAddr, destAddr, basePath)

	err = replicatePath(srcClient, destClient, basePath, "")
	if err != nil {
		return fmt.Errorf("error during replication: %v", err)
	}

	log.Println("Replication completed successfully.")
	return nil
}

func createVaultClient(address, token string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	log.Printf("Vault client created for address %s", address)
	return client, nil
}

func listSecrets(client *api.Client, path string) ([]string, error) {
	log.Printf("Listing secrets at path: %s", path)
	secret, err := client.Logical().List("secret/metadata/" + path)
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

func readSecret(client *api.Client, path string) (map[string]interface{}, error) {
	log.Printf("Reading secret at path: %s", path)
	fullPath := strings.Replace("secret/metadata/"+path, "metadata/", "data/", 1)
	secret, err := client.Logical().Read(fullPath)
	if err != nil {
		log.Printf("Error reading secret at path %s: %v", fullPath, err)
		return nil, err
	}
	if secret == nil || secret.Data["data"] == nil {
		log.Printf("No data found at path %s", fullPath)
		return nil, fmt.Errorf("no data found at path: %s", fullPath)
	}
	log.Printf("Read secret data: %v", secret.Data["data"])
	return secret.Data["data"].(map[string]interface{}), nil
}

func writeSecret(client *api.Client, path string, data map[string]interface{}) error {
	fullPath := "secret/data/" + path
	log.Printf("Writing secret to path: %s with data: %v", fullPath, data)
	_, err := client.Logical().Write(fullPath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		log.Printf("Error writing secret to path %s: %v", fullPath, err)
		return err
	}
	log.Printf("Successfully wrote secret to path: %s", fullPath)
	return nil
}

func replicatePath(srcClient, destClient *api.Client, srcPath, destPath string) error {
	log.Printf("Replicating path: %s to %s", srcPath, destPath)
	secrets, err := listSecrets(srcClient, srcPath)
	if err != nil {
		log.Printf("Error listing secrets at %s: %v", srcPath, err)
		return err
	}

	for _, secretKey := range secrets {
		if strings.HasSuffix(secretKey, "/") {
			newSrcPath := srcPath + secretKey
			newDestPath := destPath + secretKey
			err := replicatePath(srcClient, destClient, newSrcPath, newDestPath)
			if err != nil {
				log.Printf("Error replicating subpath %s: %v", newSrcPath, err)
			}
		} else {
			fullSrcPath := srcPath + secretKey
			fullDestPath := destPath + secretKey

			secretData, err := readSecret(srcClient, fullSrcPath)
			if err != nil {
				log.Printf("Error reading secret at %s: %v", fullSrcPath, err)
				continue
			}

			err = writeSecret(destClient, fullDestPath, secretData)
			if err != nil {
				log.Printf("Error writing secret to %s: %v", fullDestPath, err)
			} else {
				log.Printf("Successfully wrote secret to %s", fullDestPath)
			}
		}
	}
	return nil
}
