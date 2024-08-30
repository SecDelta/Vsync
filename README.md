<img src="./img/vsync-logo.png" alt="Vsync Logo" width="200"/>

# Vsync

![Build Status](https://github.com/SecDelta/Vsync/actions/workflows/release.yml/badge.svg)
[![Dependency Security Scan](https://github.com/SecDelta/Vsync/actions/workflows/sbom.yml/badge.svg)](https://github.com/SecDelta/Vsync/actions/workflows/sbom.yml)


A command-line tool to replicate KV Secrets between two HashiCorp Vault instances.

## Use Case

HashiCorp Vault's open-source (free) edition does not support runtime or live replication of data across Vault instances, therefore this CLI can be running as a Cronjob to implement this feature for only KV-v2 Engine.

## Features

- **Replicates Secrets**: Seamlessly replicates secrets from a source Vault instance to a destination Vault instance.
- **KV v2 Support**: Fully supports KV version 2, including handling of `metadata/` and `data/` paths.
- **Configurable Paths**: Allows replication of secrets from any specified base path within the source Vault.
- **Logging**: Provides detailed logs for tracking replication progress and troubleshooting.

## Prerequisites

- **HashiCorp Vault**: Running instances of Vault with KV version 2 enabled.

### Environment Variables

Environment Variables required to be set for running the CLI if not passed as flags to the CLI itself

| Environment Variable | Description |
|----------------------|-------------|
| SRC_VAULT_TOKEN	   | The token used to authenticate with the source Vault. |
| DEST_VAULT_TOKEN	   | The token used to authenticate with the destination Vault. |


## Usage

```sh
vsync
A Fast and Flexible Vault secrets replicator built with love by Go. It helps in implementing DR for Vault.

Usage:
  Vsync [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  kv          Replicate KV secrets from one Vault to another

Flags:
  -h, --help      help for Vsync
  -v, --version   version for Vsync

Use "Vsync [command] --help" for more information about a command.
```

```sh
vsync kv
Usage:
  Vsync kv [flags]

Flags:
      --dest-token string   Destination Vault token
  -d, --dest-vault string   Destination Vault address (required)
  -h, --help                help for kv
  -p, --path string         KV engine path (e.g., 'secret') (default "secret")
      --src-token string    Source Vault token
  -s, --src-vault string    Source Vault address (required)
```

## Testing locally

```sh
# Create two Vault instances using the docker-compose file
docker-compose up -d

# Create a secret to one of the vault instances after adding the VAULT_ADDR and VAULT_TOKEN
vault login
vault kv put secret/test password=test111

# Download the release matching your OS
# Execute the Vsync CLI 
vsync kv -s http://localhost:8200 -d http://localhost:8201 -src-token vault1 --dest-token vault2

# Check the secret in the second instance after adding the VAULT_ADDR and VAULT_TOKEN
vault login
vault kv get secert/test
```

## Future Features

- Adding support for additional Secret Engines (AWS, SSH ..etc)
