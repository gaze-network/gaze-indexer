<!-- omit from toc -->
# Gaze Indexer

Gaze Indexer is an open-source and modular indexing client for Bitcoin meta-protocols. It has support for Runes out of the box, with **Unified Consistent APIs** across fungible token protocols.

Gaze Indexer is built with **modularity** in mind, allowing users to run all modules in one monolithic instance with a single command, or as a distributed cluster of micro-services.

Gaze Indexer serves as a foundation for building ANY meta-protocol indexers, with efficient data fetching, reorg detection, and database migration tool.
This allows developers to focus on what **truly** matters: Meta-protocol indexing logic. New meta-protocols can be easily added by implementing new modules.

- [Modules](#modules)
  - [1. Runes](#1-runes)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
    - [1. Hardware Requirements](#1-hardware-requirements)
    - [2. Prepare Bitcoin Core RPC server.](#2-prepare-bitcoin-core-rpc-server)
    - [3. Prepare database.](#3-prepare-database)
    - [4. Prepare `config.yaml` file.](#4-prepare-configyaml-file)
  - [Install with Docker (recommended)](#install-with-docker-recommended)
  - [Install from source](#install-from-source)

## Modules

### 1. Runes

The Runes Indexer is our first meta-protocol indexer. It indexes Runes states, transactions, runestones, and balances using Bitcoin transactions.
It comes with a set of APIs for querying historical Runes data. See our [API Reference](https://documenter.getpostman.com/view/28396285/2sA3Bn7Cxr) for full details.

## Installation

### Prerequisites

#### 1. Hardware Requirements

Each module requires different hardware requirements.
| Module | CPU       | RAM  |
| ------ | --------- | ---- |
| Runes  | 0.5 cores | 1 GB |

#### 2. Prepare Bitcoin Core RPC server.

Gaze Indexer needs to fetch transaction data from a Bitcoin Core RPC, either self-hosted or using managed providers like QuickNode.
To self host a Bitcoin Core, see https://bitcoin.org/en/full-node.

#### 3. Prepare database.

Gaze Indexer has first-class support for PostgreSQL. If you wish to use other databases, you can implement your own database repository that satisfies each module's Data Gateway interface.
Here is our minimum database disk space requirement for each module.
| Module | Database Storage (current) | Database Storage (in 1 year) |
| ------ | -------------------------- | ---------------------------- |
| Runes  | 10 GB                      | 150 GB                       |

Here is our minimum database disk space requirement for each module.

#### 4. Prepare `config.yaml` file.

```yaml
# config.yaml
logger:
  output: TEXT # Output format for logs. current supported formats: "TEXT" | "JSON" | "GCP"
  debug: false

# Network to run the indexer on. Current supported networks: "mainnet" | "testnet"
network: mainnet

# Bitcoin Core RPC configuration options.
bitcoin_node:
  host: "" # [Required] Host of Bitcoin Core RPC (without https://)
  user: "" # Username to authenticate with Bitcoin Core RPC
  pass: "" # Password to authenticate with Bitcoin Core RPC
  disable_tls: false # Set to true to disable tls

# Block reporting configuration options. See Block Reporting section for more details.
reporting:
  disabled: false # Set to true to disable block reporting to Gaze Network. Default is false.
  base_url: "https://indexer.api.gaze.network" # Defaults to "https://indexer.api.gaze.network" if left empty
  name: "" # [Required if not disabled] Name of this indexer to show on the Gaze Network dashboard
  website_url: "" # Public website URL to show on the dashboard. Can be left empty.
  indexer_api_url: "" # Public url to access this indexer's API. Can be left empty if you want to keep your indexer private.

# HTTP server configuration options.
http_server:
  port: 8080 # Port to run the HTTP server on for modules with HTTP API handlers.

# Meta-protocol modules configuration options.
modules:
  # Configuration options for Runes module. Can be removed if not used.
  runes:
    database: "postgres" # Database to store Runes data. current supported databases: "postgres"
    datasource: "database" # Data source to be used for Bitcoin data. current supported data sources: "bitcoin-node".
    api_handlers: # API handlers to enable. current supported handlers: "http"
      - http
    postgres:
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "password"
      db_name: "postgres"
      # url: "postgres://postgres:password@localhost:5432/postgres?sslmode=prefer" # [Optional] This will override other database credentials above.
```

### Install with Docker (recommended)

We will be using `docker-compose` for our installation guide. Make sure the `docker-compose.yaml` file is in the same directory as the `config.yaml` file.

```yaml
# docker-compose.yaml
services:
  gaze-indexer:
    image: ghcr.io/gaze-network/gaze-indexer:v1.0.0
    container_name: gaze-indexer
    restart: unless-stopped
    ports:
      - 8080:8080 # Expose HTTP server port to host
    volumes:
      - "./config.yaml:/app/config.yaml" # mount config.yaml file to the container as "/app/config.yaml"
    command: ["/app/main", "run", "--runes"] # Put module flags after "run" commands to select which modules to run.
```

### Install from source

1. Install `go` version 1.22 or higher. See Go installation guide [here](https://go.dev/doc/install).
2. Clone this repository.

```bash
git clone https://github.com/gaze-network/gaze-indexer.git
cd gaze-indexer
```

3. Build the main binary.

```bash
# Get dependencies
go mod download

# Build the main binary
go build -o gaze main.go
```

4. Run database migrations with the `migrate` command and module flags.

```bash
./gaze migrate up --runes --database postgres://postgres:password@localhost:5432/postgres
```

5. Start the indexer with the `run` command and module flags.

```bash
./gaze run --runes
```

If `config.yaml` is not located at `./app/config.yaml`, use the `--config` flag to specify the path to the `config.yaml` file.

```bash
./gaze run --runes --config /path/to/config.yaml
```
