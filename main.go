//
// Copyright 2023 Kusari, Inc.
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
	"github.com/guacsec/guac/pkg/logging"
	"golang.org/x/oauth2/clientcredentials"
)

// This application utilizes oauth client credentials flow to obtain a jwt
// which can be used to create a presigned url to upload files to an authorized
// S3 bucket. Before the files get uploaded, they are converted to processor.Document
// types that GUAC understands and can ingest.
func main() {
	ctx := logging.WithLogger(context.Background())
	logger := logging.FromContext(ctx)

	if len(os.Args) != 6 {
		logger.Fatalf("Invalid args")
	}

	filePath := os.Args[1]
	clientID := os.Args[2]
	clientSecret := os.Args[3]
	tenantEndPoint := os.Args[4]
	tokenEndPoint := os.Args[5]

	authorizedClient := getAuthorizedClient(ctx, clientID, clientSecret, tokenEndPoint)

	// Check if the provided path is a directory or a file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	if fileInfo.IsDir() {
		if err := uploadDirectory(authorizedClient, tenantEndPoint, filePath); err != nil {
			fmt.Println("Error uploading:", err)
			return
		}
	} else {
		if err := uploadSingleFile(authorizedClient, tenantEndPoint, filePath); err != nil {
			fmt.Println("Error uploading:", err)
			return
		}
	}
	fmt.Println("Upload completed successfully")
}

func getAuthorizedClient(ctx context.Context, clientID, clientSecret, tokenURL string) *http.Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}

	return config.Client(ctx)
}

func getPresignedUrl(authenticatedClient *http.Client, tenantApiEndpoint string, payloadBytes []byte) (string, error) {
	resp, err := authenticatedClient.Post(tenantApiEndpoint+"/presign", "application/json", bytes.NewBuffer(payloadBytes))
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

func uploadDirectory(authenticatedClient *http.Client, tenantApiEndpoint, dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = uploadSingleFile(authenticatedClient, tenantApiEndpoint, path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func uploadSingleFile(authenticatedClient *http.Client, tenantApiEndpoint, filePath string) error {
	// Prepare the payload for the presigned URL request
	payload := map[string]string{
		"filename": filePath,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error creating JSON payload: %w", err)
	}
	presignedUrl, err := getPresignedUrl(authenticatedClient, tenantApiEndpoint, payloadBytes)
	if err != nil {
		return err
	}

	return uploadFile(presignedUrl, filePath)
}

func uploadFile(presignedUrl, filePath string) error {
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
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", body)
	}

	return nil
}
