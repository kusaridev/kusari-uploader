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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

// DocumentWrapper holds extra fields without modifying processor.Document
type DocumentWrapper struct {
	*processor.Document
	UploadMetaData *map[string]string `json:"upload_metadata,omitempty"`
}

// This application utilizes oauth client credentials flow to obtain a jwt
// which can be used to create a presigned url to upload files to an authorized
// S3 bucket. Before the files get uploaded, they are converted to processor.Document
// types that GUAC understands and can ingest.
func main() {
	// Create the root command
	var rootCmd = &cobra.Command{
		Use:   "file-uploader",
		Short: "Upload files to an S3 bucket using OAuth client credentials",
		Run:   uploadFiles,
	}

	// Define flags (new flags are optional)
	rootCmd.Flags().StringP("file-path", "f", "", "Path to file or directory to upload (required)")
	rootCmd.Flags().StringP("client-id", "c", "", "OAuth client ID (required)")
	rootCmd.Flags().StringP("client-secret", "s", "", "OAuth client secret (required)")
	rootCmd.Flags().StringP("tenant-endpoint", "t", "", "Kusari Tenant endpoint URL (required)")
	rootCmd.Flags().StringP("token-endpoint", "k", "", "Token endpoint URL (required)")
	rootCmd.Flags().StringP("alias", "a", "", "Alias that supersedes the subject in Kusari platform (optional)")
	rootCmd.Flags().StringP("document-type", "d", "", "Type of the document (image or build) sbom (optional)")

	// Bind flags to Viper with error handling
	mustBindPFlag(rootCmd, "file-path", "file-path")
	mustBindPFlag(rootCmd, "client-id", "client-id")
	mustBindPFlag(rootCmd, "client-secret", "client-secret")
	mustBindPFlag(rootCmd, "tenant-endpoint", "tenant-endpoint")
	mustBindPFlag(rootCmd, "token-endpoint", "token-endpoint")
	mustBindPFlag(rootCmd, "alias", "alias")
	mustBindPFlag(rootCmd, "document-type", "document-type")

	// Allow environment variables
	viper.SetEnvPrefix("UPLOADER")
	viper.AutomaticEnv()

	// Mark flags as required with error handling
	mustMarkFlagRequired(rootCmd, "file-path")
	mustMarkFlagRequired(rootCmd, "client-id")
	mustMarkFlagRequired(rootCmd, "client-secret")
	mustMarkFlagRequired(rootCmd, "tenant-endpoint")
	mustMarkFlagRequired(rootCmd, "token-endpoint")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

// Helper function to bind Viper flags with error handling
func mustBindPFlag(cmd *cobra.Command, configKey string, flagName string) {
	if err := viper.BindPFlag(configKey, cmd.Flags().Lookup(flagName)); err != nil {
		log.Fatal().
			Err(err).
			Str("configKey", configKey).
			Str("flagName", flagName).
			Msg("Failed to bind flag to configuration")
	}
}

// Helper function to mark flags as required with error handling
func mustMarkFlagRequired(cmd *cobra.Command, flagName string) {
	if err := cmd.MarkFlagRequired(flagName); err != nil {
		log.Fatal().
			Err(err).
			Str("flagName", flagName).
			Msg("Failed to mark flag as required")
	}
}

func uploadFiles(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Retrieve configuration values
	filePath := viper.GetString("file-path")
	clientID := viper.GetString("client-id")
	clientSecret := viper.GetString("client-secret")
	tenantEndPoint := viper.GetString("tenant-endpoint")
	tokenEndPoint := viper.GetString("token-endpoint")
	alias := viper.GetString("alias")
	docType := viper.GetString("document-type")

	// Validate required configuration
	if filePath == "" || clientID == "" || clientSecret == "" ||
		tenantEndPoint == "" || tokenEndPoint == "" {
		log.Fatal().Msg("All required parameters must be provided")
	}
	// Get authorized client
	authorizedClient := getAuthorizedClient(ctx, clientID, clientSecret, tokenEndPoint)
	defaultClient := &http.Client{}

	// Check if path is a directory or file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error getting file info")
	}

	uploadMeta := map[string]string{}
	if alias != "" {
		uploadMeta["alias"] = alias
	}
	if docType != "" {
		uploadMeta["type"] = docType
	}

	// Upload based on file type
	if fileInfo.IsDir() {
		if err := uploadDirectory(authorizedClient, defaultClient, tenantEndPoint, filePath, uploadMeta); err != nil {
			log.Fatal().
				Err(err).
				Msg("Directory upload failed")
		}
	} else {
		if err := uploadSingleFile(authorizedClient, defaultClient, tenantEndPoint, filePath, uploadMeta); err != nil {
			log.Fatal().
				Err(err).
				Msg("Single file upload failed")
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

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", fmt.Errorf("getPresignedUrl failed with unauthorized request: %d", resp.StatusCode)
		}
		// otherwise return an error
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

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
func uploadDirectory(authorizedClient, defaultClient HttpClient, tenantApiEndpoint,
	dirPath string, uploadMeta map[string]string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = uploadSingleFile(authorizedClient, defaultClient, tenantApiEndpoint, path, uploadMeta)
			if err != nil {
				return fmt.Errorf("uploadSingleFile failed with error: %w", err)
			}
		}
		return nil
	})
	return err
}

// uploadSingleFile creates a presigned URL for the filepath and calls uploadFile to upload the actual file
func uploadSingleFile(authorizedClient, defaultClient HttpClient, tenantApiEndpoint, filePath string, uploadMeta map[string]string) error {
	// check that the file is not empty
	checkFile, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get stats on filepath: %s, with error: %w", filePath, err)
	}
	// if file is empty, do not upload and return nil
	if checkFile.Size() == 0 {
		return nil
	}

	blob, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %s, err: %w", filePath, err)
	}

	// Prepare the payload for the presigned URL request
	payload := map[string]string{
		"filename": events.GetDocRef(blob),
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
	return uploadBlob(defaultClient, presignedUrl, filePath, blob, uploadMeta)
}

// uploadBlob takes the file and creates a `processor.Document` blob which is uploaded to S3
func uploadBlob(defaultClient HttpClient, presignedUrl, filePath string, readFile []byte, uploadMeta map[string]string) error {
	baseDoc := &processor.Document{
		Blob:   readFile,
		Type:   processor.DocumentUnknown,
		Format: processor.FormatUnknown,
		SourceInformation: processor.SourceInformation{
			Collector:   "Kusari-Uploader",
			Source:      fmt.Sprintf("file:///%s", filePath),
			DocumentRef: events.GetDocRef(readFile),
		},
	}

	var docByte []byte
	var err error

	if len(uploadMeta) != 0 {

		// Wrap it with additional metadata about the project
		docWrapper := DocumentWrapper{
			Document:       baseDoc,
			UploadMetaData: &uploadMeta,
		}

		docByte, err = json.Marshal(docWrapper)
		if err != nil {
			return fmt.Errorf("failed marshal of document: %w", err)
		}
	} else {
		docByte, err = json.Marshal(baseDoc)
		if err != nil {
			return fmt.Errorf("failed marshal of document: %w", err)
		}
	}

	req, err := http.NewRequest(http.MethodPut, presignedUrl, bytes.NewBuffer(docByte))
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
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("uploadBlob failed with unauthorized request: %d", resp.StatusCode)
		}
		// otherwise return an error
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", body)
	}

	return nil
}
