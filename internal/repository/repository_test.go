package repository_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/iamhectorsosa/octomap/internal/entity"
	"github.com/iamhectorsosa/octomap/internal/repository"
	"github.com/stretchr/testify/assert"
)

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
			expectedPaths: []string{"mapping: dir/file1.txt", "mapping: dir/file2.go"},
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
			expectedPaths:  []string{"mapping: file1.txt"},
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
			expectedPaths:  []string{"mapping: file1.txt"},
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
			expectedPaths:  []string{"mapping: file2.go", "mapping: file3.go"},
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

			ch := make(chan entity.Update, len(tt.files))
			done := make(chan struct{})

			var actualPaths []string

			go func() {
				defer close(done)
				for path := range ch {
					actualPaths = append(actualPaths, path.Description)
				}
			}()

			cfg := &entity.Config{
				Dir:     tt.dir,
				Include: tt.include,
				Exclude: tt.exclude,
			}

			data, err := repository.TarballReader(cfg, r, ch, 0)
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
