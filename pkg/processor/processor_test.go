package processor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gw)

		files := map[string]string{
			"repo-main/file1.go":  "package main",
			"repo-main/file2.txt": "hello world",
		}

		for name, content := range files {
			hdr := &tar.Header{
				Name: name,
				Mode: 0600,
				Size: int64(len(content)),
			}
			require.NoError(t, tw.WriteHeader(hdr))
			_, err := tw.Write([]byte(content))
			require.NoError(t, err)
		}

		require.NoError(t, tw.Close())
		require.NoError(t, gw.Close())
		w.Write(buf.Bytes())
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "process-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		config      *Config
		name        string
		wantFiles   []string
		wantUpdates int
		wantErr     bool
	}{
		{
			name: "successful processing",
			config: &Config{
				Repo:    "test-repo",
				Url:     server.URL,
				Dir:     "repo-main",
				Output:  tmpDir,
				Include: []string{".go"},
			},
			wantErr:     false,
			wantUpdates: 5, // download + mapping + 2 stats + save updates
			wantFiles:   []string{"file1.go"},
		},
		{
			name: "no file filtering",
			config: &Config{
				Repo:   "test-repo",
				Url:    server.URL,
				Dir:    "repo-main",
				Output: tmpDir,
			},
			wantErr:     false,
			wantUpdates: 6, // download + 2 mappings + 2 stats + save updates
			wantFiles:   []string{"file1.go", "file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCh := make(chan Update)
			processor := New(tt.config, updateCh)

			// Collect updates in background
			var updates []Update
			done := make(chan bool)
			go func() {
				for update := range updateCh {
					updates = append(updates, update)
				}
				done <- true
			}()

			// Run processor
			_, err := processor.Process(time.Millisecond)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			<-done

			assert.Equal(t, tt.wantUpdates, len(updates))

			files, err := filepath.Glob(filepath.Join(tmpDir, "*.json"))
			require.NoError(t, err)
			assert.Equal(t, 1, len(files))

			content, err := os.ReadFile(files[0])
			require.NoError(t, err)

			var data map[string]interface{}
			err = json.Unmarshal(content, &data)
			require.NoError(t, err)

			for _, file := range tt.wantFiles {
				assert.Contains(t, data, file)
			}
		})
	}
}
