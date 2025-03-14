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

```bash
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

### Environment Variables

You can also configure the uploader using environment variables:

```bash
export UPLOADER_FILE_PATH=/path/to/files
export UPLOADER_CLIENT_ID=your-client-id
export UPLOADER_CLIENT_SECRET=your-client-secret
export UPLOADER_TENANT_ENDPOINT=https://tenant-endpoint
export UPLOADER_TOKEN_ENDPOINT=https://token-endpoint
export UPLOADER_ALIAS="package alias"
export UPLOADER_DOCUMENT-TYPE="image"

./kusari-uploader
```

## Configuration Parameters

| Flag/Env Variable | Description | Required |
|------------------|-------------|----------|
| `-f` / `UPLOADER_FILE_PATH` | Path to file or directory to upload | Yes |
| `-c` / `UPLOADER_CLIENT_ID` | OAuth2 Client ID | Yes |
| `-s` / `UPLOADER_CLIENT_SECRET` | OAuth2 Client Secret | Yes |
| `-t` / `UPLOADER_TENANT_ENDPOINT` | Kusari Tenant endpoint URL | Yes |
| `-k` / `UPLOADER_TOKEN_ENDPOINT` | Token endpoint URL | Yes |
| `--alias` / `UPLOADER_ALIAS` | Alias that supersedes the subject in Kusari platform (optional) | No |
| `--document-type` / `UPLOADER_DOCUMENT-TYPE` | Type of the document (image or build) sbom (optional) | No |

## Help

To see all available commands and flags:

```bash
./kusari-uploader --help
```
