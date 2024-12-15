package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		expectedConfig *config
		name           string
		args           []string
		expectedErr    bool
	}{
		{
			name: "No optional flags, default branch",
			args: []string{"cmd", "user/repo"},
			expectedConfig: &config{
				repo:    "repo",
				dir:     "repo-main",
				branch:  "main",
				url:     "https://github.com/user/repo/archive/refs/heads/main.tar.gz",
				include: nil,
				exclude: nil,
				output:  "",
			},
			expectedErr: false,
		},
		{
			name:           "No optional flags, bad user/repo",
			args:           []string{"cmd", "userrepo"},
			expectedConfig: nil,
			expectedErr:    true,
		},
		{
			name: "Valid arguments with all flags",
			args: []string{"cmd", "user/repo", "--dir", "src", "--branch", "develop", "--include", ".go,.proto", "--exclude", ".mod,.sum", "--output", "~/documents"},
			expectedConfig: &config{
				repo:    "repo",
				dir:     "repo-develop/src",
				branch:  "develop",
				url:     "https://github.com/user/repo/archive/refs/heads/develop.tar.gz",
				include: []string{".go", ".proto"},
				exclude: []string{".mod", ".sum"},
				output:  "/Users/hectorsosa/documents",
			},
			expectedErr: false,
		},
		{
			name: "Empty include and exclude flags",
			args: []string{"cmd", "user/repo", "--include", "", "--exclude", ""},
			expectedConfig: &config{
				repo:    "repo",
				dir:     "repo-main",
				branch:  "main",
				url:     "https://github.com/user/repo/archive/refs/heads/main.tar.gz",
				include: nil,
				exclude: nil,
				output:  "",
			},
			expectedErr: false,
		},
		{
			name:           "Missing repository argument",
			args:           []string{"cmd"},
			expectedConfig: nil,
			expectedErr:    true,
		},
		{
			name:           "Error in flag parsing",
			args:           []string{"cmd", "user/repo", "--unknown", "value"},
			expectedConfig: nil,
			expectedErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = tt.args
			cfg, err := getConfig()

			if !tt.expectedErr {
				assert.NotNil(t, cfg, "Expected config to be non-nil")
				assert.NoError(t, err, "Expected no error, but got one")
				assert.Equal(t, tt.expectedConfig, cfg, "Config mismatch")
			} else {
				assert.Error(t, err, "Expected an error, but got none")
				assert.Nil(t, cfg, "Expected config to be nil")
			}
		})
	}
}

func TestTarballReader(t *testing.T) {
	tests := []struct {
		files          map[string]string
		expectedData   map[string]interface{}
		name           string
		dir            string
		include        []string
		exclude        []string
		expectedPaths  []string
		expectedErrors []string
	}{
		{
			name: "Test no dir no filters",
			files: map[string]string{
				"dir/file1.txt": "Content 1",
				"dir/file2.go":  "Content 2",
			},
			dir:           "",
			include:       nil,
			exclude:       nil,
			expectedPaths: []string{"dir/file1.txt", "dir/file2.go"},
			expectedData: map[string]interface{}{
				"dir": map[string]interface{}{
					"file1.txt": "Content 1",
					"file2.go":  "Content 2",
				},
			},
		},
		{
			name: "Test include filter",
			files: map[string]string{
				"dir/file1.txt": "Content 1",
				"dir/file2.go":  "Content 2",
			},
			dir:            "dir",
			include:        []string{".txt"},
			exclude:        nil,
			expectedPaths:  []string{"file1.txt"},
			expectedErrors: nil,
			expectedData: map[string]interface{}{
				"file1.txt": "Content 1",
			},
		},
		{
			name: "Test exclude filter",
			files: map[string]string{
				"dir/file1.txt": "Content 1",
				"dir/file2.go":  "Content 2",
			},
			dir:            "dir",
			include:        nil,
			exclude:        []string{".go"},
			expectedPaths:  []string{"file1.txt"},
			expectedErrors: nil,
			expectedData: map[string]interface{}{
				"file1.txt": "Content 1",
			},
		},
		{
			name: "Test include and exclude filter",
			files: map[string]string{
				"dir/file1.txt": "Content 1",
				"dir/file2.go":  "Content 2",
				"dir/file3.go":  "Content 3",
			},
			dir:            "dir",
			include:        []string{".go", ".txt"},
			exclude:        []string{"file1.txt"},
			expectedPaths:  []string{"file2.go", "file3.go"},
			expectedErrors: nil,
			expectedData: map[string]interface{}{
				"file2.go": "Content 2",
				"file3.go": "Content 3",
			},
		},
		{
			name: "Test no files match",
			files: map[string]string{
				"dir/file1.txt": "Content 1",
				"dir/file2.go":  "Content 2",
			},
			dir:            "dir",
			include:        []string{".md"},
			exclude:        nil,
			expectedPaths:  nil,
			expectedErrors: nil,
			expectedData:   map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := createTestTarball(tt.files)
			if err != nil {
				t.Fatalf("Failed to create test tarball: %v", err)
			}

			ch := make(chan interface{}, len(tt.files))
			done := make(chan struct{})

			var actualPaths []string

			go func() {
				defer close(done)
				for path := range ch {
					actualPaths = append(actualPaths, path.(string))
				}
			}()

			data, err := tarballReader(tt.dir, tt.include, tt.exclude, r, ch, 0)
			close(ch)

			if err != nil {
				t.Fatalf("Error in tarballReader: %v", err)
			}

			<-done

			assert.ElementsMatch(t, tt.expectedPaths, actualPaths,
				"Paths should match expected paths")

			assert.Equal(t, tt.expectedData, data,
				"Data should match expected data")
		})
	}
}

func createTestTarball(files map[string]string) (io.Reader, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for filename, content := range files {
		hdr := &tar.Header{
			Name: filename,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}
