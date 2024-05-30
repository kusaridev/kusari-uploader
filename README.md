# kusari-uploader

This application is used to ingests SBOMs and Attestations into
[GUAC](https://github.com/guacsec/guac). It uses oauth2's client credential flow to generate a pre-signed URL to allow for uploads to an authorized S3.
This will allow for SBOMs and other documents to be uploaded, converted to `processor.Document` blob and be uploaded to S3 without the need to use
`guacone` or `guaccollect`. Further functionality can be added to ingest other information from the source if needed.

## Inputs

### `files`

**Required** Path to directory or specific file to ingest

### `clientID`

**Required** OAuth2 Client Id

### `clientSecret`

**Required** OAuth2 Client Secret

### `tenantEndPoint`

**Required** URL for the specific tenant

### `tokenEndPoint`

**Required** URL of auth token provider

## Outputs

### `console_out`

Success message
