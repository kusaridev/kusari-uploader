# kusari-uploader

This application ingests SBOMs and Attestations into the Kusari Platform. It uses OAuth2's client credentials flow to generate a pre-signed URL for uploads to an authorized S3 bucket.

## Features

-   Upload single files or entire directories
-   OAuth2 client credentials authentication
-   Flexible configuration via CLI flags or environment variables
-   Supports ingestion of SBOMs and other documents
-   Optional metadata tagging of uploaded documents

## Installation

```bash
go get github.com/kusaridev/kusari-uploader

## Usage

### Command-Line Flags

# Upload a single file
./kusari-uploader -f /path/to/file \
    -c CLIENT_ID \
    -s CLIENT_SECRET \
    -t TENANT_ENDPOINT \
    -k TOKEN_ENDPOINT \
    -a "package alias" \
    -d "image"

# Upload an entire directory
./kusari-uploader -f /path/to/directory \
    -c CLIENT_ID \
    -s CLIENT_SECRET \
    -t TENANT_ENDPOINT \
    -k TOKEN_ENDPOINT \
    -a "package alias" \
    -d "image"
```

## Configuration Parameters

| Short Flag/ Full Flag | Description | Required |
|------------------|-------------|----------|
| `-f` / `--file-path` | Path to file or directory to upload | Yes |
| `-c` / `--client-id` | OAuth2 Client ID | Yes |
| `-s` / `--client-secret` | OAuth2 Client Secret | Yes |
| `-t` / `--tenant-endpoint` | Kusari Tenant endpoint URL | Yes |
| `-k` / `--token-endpoint` | Token endpoint URL | Yes |
| `--alias` / `--alias` | Alias that supersedes the subject in Kusari platform (optional) | No |
| `--document-type` / `--document-type` | Type of the document (image or build) sbom (optional) | No |

## Help

To see all available commands and flags:

```bash
./kusari-uploader --help
Usage:
  file-uploader [flags]

Flags:
  -a, --alias string             Alias that supersedes the subject in Kusari platform (optional)
  -c, --client-id string         OAuth client ID (required)
  -s, --client-secret string     OAuth client secret (required)
  -d, --document-type string     Type of the document (image or build) sbom (optional)
  -f, --file-path string         Path to file or directory to upload (required)
  -h, --help                     help for file-uploader
  -t, --tenant-endpoint string   Kusari Tenant endpoint URL (required)
  -k, --token-endpoint string    Token endpoint URL (required)
```
