//
// Copyright 2024 Kusari, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/guacsec/guac/pkg/events"
	"github.com/guacsec/guac/pkg/handler/processor"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2/clientcredentials"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

// This application utilizes oauth client credentials flow to obtain a jwt
// which can be used to create a presigned url to upload files to an authorized
// S3 bucket. Before the files get uploaded, they are converted to processor.Document
// types that GUAC understands and can ingest.
func main() {
	ctx := context.Background()

	if len(os.Args) != 6 {
		log.Fatal().Msg("Invalid args")
	}

	filePath := os.Args[1]
	clientID := os.Args[2]
	clientSecret := os.Args[3]
	tenantEndPoint := os.Args[4]
	tokenEndPoint := os.Args[5]

	authorizedClient := getAuthorizedClient(ctx, clientID, clientSecret, tokenEndPoint)

	// check if the provided path is a directory or a file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("error getting file info:")
	}

	defaultClient := &http.Client{}

	if fileInfo.IsDir() {
		if err := uploadDirectory(authorizedClient, defaultClient, tenantEndPoint, filePath); err != nil {
			log.Fatal().
				Err(err).
				Msg("uploadDirectory failed with error")
		}
	} else {
		if err := uploadSingleFile(authorizedClient, defaultClient, tenantEndPoint, filePath); err != nil {
			log.Fatal().
				Err(err).
				Msg("uploadSingleFile failed with error")
		}
	}
	fmt.Println("Upload completed successfully")
}

// getAuthorizedClient utilizes oauth2 client credential flow to obtain an authorized client
func getAuthorizedClient(ctx context.Context, clientID, clientSecret, tokenURL string) HttpClient {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}

	return config.Client(ctx)
}

// getPresignedUrl utilizes authorized client to obtain the presigned URL to upload to S3
func getPresignedUrl(authorizedClient HttpClient, tenantApiEndpoint string, payloadBytes []byte) (string, error) {
	resp, err := authorizedClient.Post(tenantApiEndpoint+"/presign", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to POST to tenant endpoint: %s, with error: %w", tenantApiEndpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body with error: %w", err)
	}

	type url struct {
		PresignedUrl string `json:"presignedUrl"`
	}

	var result url
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal the results with body: %s with error: %w", string(body), err)
	}

	presignedUrl := result.PresignedUrl

	return presignedUrl, nil
}

// uploadDirectory uses filepath.Walk to walk through the directory and upload the files that are found
func uploadDirectory(authorizedClient, defaultClient HttpClient, tenantApiEndpoint, dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = uploadSingleFile(authorizedClient, defaultClient, tenantApiEndpoint, path)
			if err != nil {
				return fmt.Errorf("uploadSingleFile failed with error: %w", err)
			}
		}
		return nil
	})
	return err
}

// uploadSingleFile creates a presigned URL for the filepath and calls uploadFile to upload the actual file
func uploadSingleFile(authorizedClient, defaultClient HttpClient, tenantApiEndpoint, filePath string) error {
	// Prepare the payload for the presigned URL request
	payload := map[string]string{
		"filename": filePath,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error creating JSON payload: %w", err)
	}
	presignedUrl, err := getPresignedUrl(authorizedClient, tenantApiEndpoint, payloadBytes)
	if err != nil {
		return err
	}

	// pass in default client without the jwt other wise it will error with both the presigned url and jwt
	return uploadBlob(defaultClient, presignedUrl, filePath)
}

// uploadBlob takes the file and creates a `processor.Document` blob which is uploaded to S3
func uploadBlob(defaultClient HttpClient, presignedUrl, filePath string) error {
	blob, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %s, err: %w", filePath, err)
	}

	doc := &processor.Document{
		Blob:   blob,
		Type:   processor.DocumentUnknown,
		Format: processor.FormatUnknown,
		SourceInformation: processor.SourceInformation{
			Collector:   "Kusari-Uploader",
			Source:      fmt.Sprintf("file:///%s", filePath),
			DocumentRef: events.GetDocRef(blob),
		},
	}

	docByte, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed marshal of document: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, presignedUrl, bytes.NewReader(docByte))
	if err != nil {
		return fmt.Errorf("failed to create new http request with error: %w", err)
	}

	req.Header.Set("Content-Type", "multipart/form-data")

	resp, err := defaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to http.Client Do with error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", body)
	}

	return nil
}
