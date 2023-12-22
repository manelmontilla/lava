// Copyright 2023 Adevinta

package urlutil

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGet_HTTP(t *testing.T) {
	tests := []struct {
		name        string
		handlerFunc func(http.ResponseWriter, *http.Request)
		want        []byte
		wantNilErr  bool
	}{
		{
			name: "valid",
			handlerFunc: func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprintln(writer, "response body")
			},
			want:       []byte("response body\n"),
			wantNilErr: true,
		},
		{
			name:        "empty",
			handlerFunc: func(writer http.ResponseWriter, request *http.Request) {},
			want:        []byte{},
			wantNilErr:  true,
		},
		{
			name: "not found",
			handlerFunc: func(writer http.ResponseWriter, request *http.Request) {
				http.Error(writer, "not found", http.StatusNotFound)
			},
			want:       nil,
			wantNilErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.handlerFunc))
			defer ts.Close()
			// Ignore the possible error parsing the url; the test server will
			// always return a valid one.
			u, _ := url.Parse(ts.URL)
			got, err := Get(u)
			if (err == nil) != tt.wantNilErr {
				t.Fatalf("unexpected error: want nil: %v, got: %v", tt.wantNilErr, err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestGet_URL(t *testing.T) {
	tests := []struct {
		name    string
		url     *url.URL
		want    []byte
		wantErr error
	}{
		{
			name:    "file",
			url:     mustParseURL("testdata/content.txt"),
			want:    []byte("file with content\n"),
			wantErr: nil,
		},
		{
			name:    "empty file",
			url:     mustParseURL("testdata/empty.txt"),
			want:    []byte{},
			wantErr: nil,
		},
		{
			name:    "file does not exist",
			url:     mustParseURL("testdata/not_exist"),
			want:    nil,
			wantErr: os.ErrNotExist,
		},
		{
			name:    "empty file path",
			url:     mustParseURL(""),
			want:    nil,
			wantErr: os.ErrNotExist,
		},
		{
			name:    "invalid scheme",
			url:     mustParseURL("invalid://example.com/file.json"),
			want:    nil,
			wantErr: ErrInvalidScheme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.url)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("unexpected error: want: %v, got: %v", tt.wantErr, err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return url
}
