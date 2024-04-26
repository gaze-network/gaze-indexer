<!-- omit from toc -->
# Gaze Indexer
Gaze Indexer is an open-source and modular indexing client for Bitcoin meta-protocols. It has support for Bitcoin and Runes out of the box, 
but other meta-protocols can be easily implemented.

- [Modules](#modules)
  - [1. Bitcoin](#1-bitcoin)
  - [2. Runes](#2-runes)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
    - [1. Prepare `config.yaml` file.](#1-prepare-configyaml-file)
    - [2. Prepare Bitcoin Core RPC server.](#2-prepare-bitcoin-core-rpc-server)
    - [3. Hardware Requirements](#3-hardware-requirements)
  - [Install with Docker (recommended)](#install-with-docker-recommended)
  - [Install from source](#install-from-source)
- [Architecture](#architecture)
- [Block Reporting](#block-reporting)
- [Implementing a new meta-protocol module](#implementing-a-new-meta-protocol-module)
- [Contributing](#contributing)

## Modules
### 1. Bitcoin
The Bitcoin Indexer, the heart of every meta-protocol, is responsible for indexing **Bitcoin transactions, blocks, and UTXOs**. It requires a Bitcoin Core RPC as source of Bitcoin transactions, 
and stores the indexed data in database to be used by other modules.

### 2. Runes
The Runes Indexer is our first meta-protocol indexer. It indexes Runes states, transactions, runestones, and balances using Bitcoin transactions.
It comes with a set of APIs for querying historical Runes data. See our [API Reference](https://documenter.getpostman.com/view/28396285/2sA3Bn7Cxr) for full details.

## Installation
### Prerequisites
#### 1. Prepare `config.yaml` file.
```yaml
# config.yaml
logger:
  output: text
  debug: false

bitcoin_node:
  host: "" # [Required] Host of Bitcoin Core RPC (without https://)
  user: "" # [Required] Username to authenticate with Bitcoin Core RPC
  pass: "" # [Required] Password to authenticate with Bitcoin Core RPC
  disable_tls: false # Set to true to disable tls

network: mainnet # Network to run the indexer on. Current supported networks: "mainnet" | "testnet"

reporting: # Block reporting configuration options. See Block Reporting section for more details.
  disabled: false # Set to true to disable block reporting to Gaze Network. Default is false.
  base_url: "https://indexer.api.gaze.network" # Defaults to "https://indexer.api.gaze.network" if left empty
  name: "" # [Required if not disabled] Name of this indexer to show on the Gaze Network dashboard
  website_url: "" # Public website URL to show on the dashboard. Can be left empty.
  indexer_api_url: "" # Public url to access this indexer's API. Can be left empty if you want to keep your indexer private.

้http_server:
  port: 8080 # Port to run the HTTP server on for modules with HTTP API handlers.

modules:
  bitcoin: # Configuration options for Bitcoin module. Can be removed if not used.
    database: "postgres" # Database to store bitcoin data. current supported databases: "postgres"
    postgres:
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "password"
      db_name: "postgres"
      # url: "postgres://postgres:password@localhost:5432/postgres?sslmode=prefer" # [Optional] This will override other database credentials above.
  runes: # Configuration options for Runes module. Can be removed if not used.
    database: "postgres" # Database to store Runes data. current supported databases: "postgres"
    datasource: "bitcoin-node" # Data source to be used for Bitcoin data. current supported data sources: "bitcoin-node" | "postgres". If "postgres" is used, the connected database must be the same database as the Bitcoin module.
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

#### 2. Prepare Bitcoin Core RPC server.  
Gaze Indexer needs to fetch transaction data from a Bitcoin Core RPC, either self-hosted or using managed providers like QuickNode.
To self host a Bitcoin Core, see https://bitcoin.org/en/full-node.

#### 3. Hardware Requirements
Each module requires different hardware requirements.
| Module  | CPU        | RAM    | Database Storage |
| ------- | ---------- | ------ | ---------------- |
| Bitcoin | 0.25 cores | 256 MB | 240 GB           |
| Runes   | 0.5 cores  | 1 GB   | 150 GB           |

### Install with Docker (recommended)
We will be using `docker-compose` for our installation guide. Make sure the `docker-compose.yaml` file is in the same directory as the `config.yaml` file.
```yaml
# docker-compose.yaml
services:
  gaze-indexer:
    image: ghcr.io/gaze-network/gaze-indexer:v1.0.0
    container_name: gaze-indexer
    restart: unless-stopped
    volumes:
      - './config.yaml:/app/config.yaml' # mount config.yaml file to the container as "/app/config.yaml"
    command: ["/app/main", "run", "--bitcoin", "--runes"] # Put module flags after "run" commands to select which modules to run.
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
4. Run the main binary with the `run` command and module flags.
```bash
./gaze run --bitcoin --runes
```
If `config.yaml` is not located at `./app/config.yaml`, use the `--config` flag to specify the path to the `config.yaml` file.
```bash
./gaze run --bitcoin --runes --config /path/to/config.yaml
```

## Architecture
TODO

## Block Reporting
Gaze Indexer has a block reporting system for verifying data integrity of indexers. Visit the [Gaze Network dashboard](https://dash.gaze.network) to see the status of other indexers.
By default, block reporting is enabled and will submit node and block reports to https://indexer.api.gaze.network.
If you want to opt-out of block reporting, set `reporting.disabled` to `true` in the `config.yaml` file.

## Implementing a new meta-protocol module
TODO

## Contributing
TODO
