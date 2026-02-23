# kusari-uploader

**kusari-uploader will reach end-of-life on April 7, 2026.**
After this date, kusari-uploader will receive no further updates, including security updates.
Users will need to switch to the [kusari-cli](https://docs.kusari.cloud/software/ingest-sboms/kusari-uploader) tool before then.
The [kusari-ingest](https://docs.kusari.cloud/software/ingest-sboms/github-action) GitHub Action now uses kusari-cli, so users of that Action do not need to make any changes.

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
| `-k` / `--token-endpoint` | Token endpoint URL | No |
| `--alias` | Alias that supersedes the subject in Kusari platform (optional) | No |
| `--document-type` | Type of the document (image or build) sbom (optional) | No |
| `--open-vex` | Indicate that this is an OpenVEX document (only works with files) | No |
| `--tag` | Tag value to set in the document wrapper upload meta (e.g. govulncheck) | No |
| `--software-id` | Kusari Platform Software ID value to set in the document wrapper upload meta | No |
| `--sbom-subject` | Kusari Platform Software sbom subject substring value to set in the document wrapper upload meta | No |
| `--component-name` | Kusari Platform component name | No |
| `--check-blocked-packages` | Check if any of the SBOMs uses a package contained in the blocked package list | No |

## Help

To see all available commands and flags:

```bash
./kusari-uploader --help
Upload files to an S3 bucket using OAuth client credentials

Usage:
  file-uploader [flags]

Flags:
  -a, --alias string             Alias that supersedes the subject in Kusari platform (optional)
      --check-blocked-packages   Check if any of the SBOMs uses a package contained in the blocked package list
  -c, --client-id string         OAuth client ID (required)
  -s, --client-secret string     OAuth client secret (required)
      --component-name string    Kusari Platform component name (optional)
  -d, --document-type string     Type of the document (image or build) sbom (optional)
  -f, --file-path string         Path to file or directory to upload (required)
  -h, --help                     help for file-uploader
      --open-vex                 Indicate that this is an OpenVEX document (optional, only works with files)
      --sbom-subject string      Kusari Platform Software sbom subject substring value to set in the document wrapper upload meta (optional)
      --software-id string       Kusari Platform Software ID value to set in the document wrapper upload meta (optional)
      --tag string               Tag value to set in the document wrapper upload meta (optional, e.g. govulncheck)
  -t, --tenant-endpoint string   Kusari Tenant endpoint URL (required)
  -k, --token-endpoint string    Token endpoint URL
  ```
