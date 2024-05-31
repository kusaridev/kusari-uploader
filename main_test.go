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
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type ClientMock struct {
	DoFunc   func(req *http.Request) (*http.Response, error)
	PostFunc func(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	if c.DoFunc != nil {
		return c.DoFunc(req)
	}
	return &http.Response{}, nil
}

func (c *ClientMock) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	if c.PostFunc != nil {
		return c.PostFunc(url, contentType, body)
	}
	return &http.Response{}, nil
}

func Test_getPresignedUrl(t *testing.T) {
	type args struct {
		authenticatedClient *ClientMock
		tenantApiEndpoint   string
		filename            string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Successful Presigned URL Retrieval",
			args: args{
				authenticatedClient: &ClientMock{
					PostFunc: func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
						returnedBody := `{"presignedUrl": "http://example.com/upload"}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(returnedBody)),
						}, nil
					},
				},
				tenantApiEndpoint: "http://example.com",
				filename:          "testfile.txt",
			},
			want:    "http://example.com/upload",
			wantErr: false,
		},
		{
			name: "Failed Presigned URL Retrieval",
			args: args{
				authenticatedClient: &ClientMock{
					PostFunc: func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
				},
				tenantApiEndpoint: "http://example.com",
				filename:          "testfile.txt",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{
				"filename": tt.args.filename,
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("error creating JSON payload: %v", err)
				return
			}
			got, err := getPresignedUrl(tt.args.authenticatedClient, tt.args.tenantApiEndpoint, payloadBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPresignedUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPresignedUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_uploadDirectory(t *testing.T) {
	authClientMock := &ClientMock{
		PostFunc: func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
			returnedBody := `{"presignedUrl": "http://example.com/upload"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(returnedBody)),
			}, nil
		},
	}

	defaultClientMock := &ClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	type args struct {
		tenantApiEndpoint string
		dirPath           string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful Directory Upload",
			args: args{
				tenantApiEndpoint: "http://example.com",
				dirPath:           "./testdata",
			},
			wantErr: false,
		},
		{
			name: "Failed Directory Upload - Non-existent Directory",
			args: args{
				tenantApiEndpoint: "http://example.com",
				dirPath:           "./nonexistent",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uploadDirectory(authClientMock, defaultClientMock, tt.args.tenantApiEndpoint, tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("uploadDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uploadSingleFile(t *testing.T) {
	type args struct {
		authenticatedClient *ClientMock
		defaultClient       *ClientMock
		tenantApiEndpoint   string
		filePath            string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful Single File Upload",
			args: args{
				authenticatedClient: &ClientMock{
					PostFunc: func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
						returnedBody := `{"presignedUrl": "http://example.com/upload"}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(returnedBody)),
						}, nil
					},
				},
				defaultClient: &ClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
				},
				tenantApiEndpoint: "http://example.com",
				filePath:          "./testdata/hello",
			},
			wantErr: false,
		},
		{
			name: "skip empty file",
			args: args{
				tenantApiEndpoint: "http://example.com",
				filePath:          "./testdata/empty",
			},
			wantErr: false,
		},
		{
			name: "Failed Single File Upload - Presigned URL Error",
			args: args{
				authenticatedClient: &ClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
					PostFunc: func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
						returnedBody := `{"presignedUrl": "http://example.com/upload"}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(returnedBody)),
						}, nil
					},
				},
				defaultClient: &ClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
				},
				tenantApiEndpoint: "http://example.com",
				filePath:          "./testdata/hello",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uploadSingleFile(tt.args.authenticatedClient, tt.args.defaultClient, tt.args.tenantApiEndpoint, tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("uploadSingleFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uploadBlob(t *testing.T) {
	type args struct {
		authenticatedClient *ClientMock
		presignedUrl        string
		filePath            string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful File Upload",
			args: args{
				authenticatedClient: &ClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
				},
				presignedUrl: "http://example.com/upload",
				filePath:     "./testdata/hello",
			},
			wantErr: false,
		},
		{
			name: "Failed File Upload - Invalid URL",
			args: args{
				authenticatedClient: &ClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewBufferString("")),
						}, nil
					},
				},
				presignedUrl: "http://invalid-url",
				filePath:     "./testdata/hello",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uploadBlob(tt.args.authenticatedClient, tt.args.presignedUrl, tt.args.filePath, []byte("hello")); (err != nil) != tt.wantErr {
				t.Errorf("uploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
