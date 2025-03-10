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
    --project-name "My Project" \
    --repo-name "My Repo" \
    --poc-name "John Doe" \
    --poc-email "john.doe@example.com"

# Upload an entire directory
./kusari-uploader -f /path/to/directory \
    -c CLIENT_ID \
    -s CLIENT_SECRET \
    -t TENANT_ENDPOINT \
    -k TOKEN_ENDPOINT \
    --project-name "My Project" \
    --repo-name "My Repo" \
    --poc-name "John Doe" \
    --poc-email "john.doe@example.com"
```

### Environment Variables

You can also configure the uploader using environment variables:

```bash
export UPLOADER_FILE_PATH=/path/to/files
export UPLOADER_CLIENT_ID=your-client-id
export UPLOADER_CLIENT_SECRET=your-client-secret
export UPLOADER_TENANT_ENDPOINT=https://tenant-endpoint
export UPLOADER_TOKEN_ENDPOINT=https://token-endpoint
export UPLOADER_PROJECT_NAME="My Project"
export UPLOADER_REPO_NAME="My Repo"
export UPLOADER_POC_NAME="John Doe"
export UPLOADER_POC_EMAIL="john.doe@example.com"

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
| `--project-name` / `UPLOADER_PROJECT_NAME` | Project Name tag to associate with project in Kusari platform | Yes |
| `--repo-name` / `UPLOADER_REPO_NAME` | Repository Name | Yes |
| `--poc-name` / `UPLOADER_POC_NAME` | Point of Contact Name | Yes |
| `--poc-email` / `UPLOADER_POC_EMAIL` | TPoint of Contact Email | Yes |

## Help

To see all available commands and flags:

```bash
./kusari-uploader --help
```
