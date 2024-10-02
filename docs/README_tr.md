## Çeviriler
- [English (İngilizce)](../README.md)

**Son Güncelleme:** 21 Ağustos 2024
> **Not:** Bu belge, topluluk tarafından yapılmış bir çeviridir. Ana README.md dosyasındaki güncellemeler buraya otomatik olarak yansıtılmayabilir. En güncel bilgiler için [İngilizce sürümü](../README.md) inceleyin.


# Gaze Indexer

Gaze Indexer, değiştirilebilir token protokolleri arasında **Birleştirilmiş Tutarlı API'lere** sahip Bitcoin meta-protokolleri için açık kaynaklı ve modüler bir indeksleme istemcisidir.

Gaze Indexer, kullanıcıların tüm modülleri tek bir komutla tek bir monolitik örnekte veya dağıtılmış bir mikro hizmet kümesi olarak çalıştırmasına olanak tanıyan **modülerlik** göz önünde bulundurularak oluşturulmuştur.

Gaze Indexer, verimli veri getirme, yeniden düzenleme algılama ve veritabanı taşıma aracı ile HERHANGİ bir meta-protokol indeksleyici oluşturmak için bir temel görevi görür.
Bu, geliştiricilerin **gerçekten** önemli olana odaklanmasını sağlar: Meta-protokol indeksleme mantığı. Yeni meta-protokoller, yeni modüller uygulanarak kolayca eklenebilir.

- [Modüller](#modules)
  - [1. Runes](#1-runes)
- [Kurulum](#installation)
  - [Önkoşullar](#prerequisites)
    - [1. Donanım Gereksinimleri](#1-hardware-requirements)
    - [2. Bitcoin Core RPC sunucusunu hazırlayın.](#2-prepare-bitcoin-core-rpc-server)
    - [3. Veritabanı hazırlayın.](#3-prepare-database)
    - [4.  `config.yaml` dosyasını hazırlayın.](#4-prepare-configyaml-file)
  - [Docker ile yükle (önerilir)](#install-with-docker-recommended)
  - [Kaynaktan yükle](#install-from-source)

## Modüller

### 1. Runes

Runes Dizinleyici ilk meta-protokol dizinleyicimizdir. Bitcoin işlemlerini kullanarak Runes durumlarını, işlemlerini, rün taşlarını ve bakiyelerini indeksler.
Geçmiş Runes verilerini sorgulamak için bir dizi API ile birlikte gelir. Tüm ayrıntılar için [API Referansı] (https://api-docs.gaze.network) adresimize bakın.


## Kurulum

### Önkoşullar

#### 1. Donanım Gereksinimleri

Her modül farklı donanım gereksinimleri gerektirir.
| Modül | CPU | RAM |
| ------ | --------- | ---- |
| Runes | 0,5 çekirdek | 1 GB |

#### 2. Bitcoin Core RPC sunucusunu hazırlayın.

Gaze Indexer'ın işlem verilerini kendi barındırdığı ya da QuickNode gibi yönetilen sağlayıcıları kullanan bir Bitcoin Core RPC'den alması gerekir.
Bir Bitcoin Core'u kendiniz barındırmak için bkz. https://bitcoin.org/en/full-node.

#### 3. Veritabanını hazırlayın.

Gaze Indexer PostgreSQL için birinci sınıf desteğe sahiptir. Diğer veritabanlarını kullanmak isterseniz, her modülün Veri Ağ Geçidi arayüzünü karşılayan kendi veritabanı havuzunuzu uygulayabilirsiniz.
İşte her modül için minimum veritabanı disk alanı gereksinimimiz.
| Modül | Veritabanı Depolama Alanı (mevcut) | Veritabanı Depolama Alanı (1 yıl içinde) |
| ------ | -------------------------- | ---------------------------- |
| Runes | 10 GB | 150 GB |

#### 4. config.yaml` dosyasını hazırlayın.

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
    datasource: "bitcoin-node" # Data source to be used for Bitcoin data. current supported data sources: "bitcoin-node".
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

### Docker ile yükleyin (önerilir)

Kurulum kılavuzumuz için `docker-compose` kullanacağız. Docker-compose.yaml` dosyasının `config.yaml` dosyası ile aynı dizinde olduğundan emin olun.

```yaml
# docker-compose.yaml
services:
  gaze-indexer:
    image: ghcr.io/gaze-network/gaze-indexer:v0.2.1
    container_name: gaze-indexer
    restart: unless-stopped
    ports:
      - 8080:8080 # Expose HTTP server port to host
    volumes:
      - "./config.yaml:/app/config.yaml" # mount config.yaml file to the container as "/app/config.yaml"
    command: ["/app/main", "run", "--modules", "runes"] # Put module flags after "run" commands to select which modules to run.
```

### Kaynaktan yükleyin

1. Go` sürüm 1.22 veya daha üstünü yükleyin. Go kurulum kılavuzuna bakın [burada](https://go.dev/doc/install).
2. Bu depoyu klonlayın.

```bash
git clone https://github.com/gaze-network/gaze-indexer.git
cd gaze-indexer
```

3. Ana ikili dosyayı oluşturun.

```bash
# Bağımlılıkları al
go mod indir

# Ana ikili dosyayı oluşturun
go build -o gaze main.go
```

4. Veritabanı geçişlerini `migrate` komutu ve modül bayrakları ile çalıştırın.

```bash
./gaze migrate up --runes --database postgres://postgres:password@localhost:5432/postgres
```

5. Dizinleyiciyi `run` komutu ve modül bayrakları ile başlatın.

```bash
./gaze run --modules runes
```

Eğer `config.yaml` dosyası `./app/config.yaml` adresinde bulunmuyorsa, `config.yaml` dosyasının yolunu belirtmek için `--config` bayrağını kullanın.

```bash
./gaze run --modules runes --config /path/to/config.yaml
```


## Çeviriler
- [English (İngilizce)](../README.md)
