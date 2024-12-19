package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		err  error
		name string
		slug string
	}{
		{
			name: "Valid slug",
			slug: "user/repo",
		},
		{
			name: "Missing slash",
			slug: "userrepo",
			err:  fmt.Errorf(invalidUserRepoTxt, "userrepo"),
		},
		{
			name: "Too many slashes",
			slug: "user/repo/another",
			err:  fmt.Errorf(invalidUserRepoTxt, "user/repo/another"),
		},
		{
			name: "Empty user",
			slug: "/repo",
			err:  fmt.Errorf(invalidUserRepoTxt, "/repo"),
		},
		{
			name: "Empty repo",
			slug: "user/",
			err:  fmt.Errorf(invalidUserRepoTxt, "user/"),
		},
		{
			name: "Empty slug",
			slug: "",
			err:  fmt.Errorf(invalidUserRepoTxt, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlug(tt.slug)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBranch(t *testing.T) {
	tests := []struct {
		err    error
		name   string
		branch string
	}{
		{
			name:   "Valid branch",
			branch: "main",
		},
		{
			name:   "Empty branch",
			branch: "",
			err:    fmt.Errorf(invalidBranchName, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranch(tt.branch)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRepoDetails(t *testing.T) {
	slug := "user/repo"
	branch := "main"
	inputDir := "src"

	expectedRepo := "repo"
	expectedURL := "https://github.com/user/repo/archive/refs/heads/main.tar.gz"
	expectedDir := "repo-main/src"

	repo, url, dir := createRepoDetails(slug, branch, inputDir)

	assert.Equal(t, expectedRepo, repo)
	assert.Equal(t, expectedURL, url)
	assert.Equal(t, expectedDir, dir)
}

func TestValidateOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		err    error
		name   string
		output string
	}{
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "valid directory path",
			output: testDir,
		},
		{
			name:   "output with extension",
			output: "test.txt",
			err:    fmt.Errorf(invalidOutputWithExt, "test.txt"),
		},
		{
			name:   "non-existent directory",
			output: filepath.Join(tmpDir, "nonexistst"),
			err: fmt.Errorf(
				errOutputDoesntExit,
				filepath.Join(tmpDir, "nonexistst"),
				fmt.Errorf("stat %s: no such file or directory", filepath.Join(tmpDir, "nonexistst")),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutput(tt.output)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolveOuput(t *testing.T) {
	origEnv := os.Getenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", origEnv)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name    string
		output  string
		envVars map[string]string
		want    string
	}{
		{
			name:   "empty output",
			output: "",
			want:   "",
		},
		{
			name:   "regular directory path",
			output: "/path/to/output",
			want:   "/path/to/output",
		},
		{
			name:    "directory path with environment variable",
			output:  "$TEST_VAR/output",
			envVars: map[string]string{"TEST_VAR": "/test/path"},
			want:    "/test/path/output",
		},
		{
			name:   "directory path with home directory",
			output: "~/output",
			want:   filepath.Join(homeDir, "output"),
		},
		{
			name:   "relative directory path without home prefix",
			output: "relative/path",
			want:   "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got, err := resolveOutput(tt.output)
			if err != nil {
				t.Errorf("failed to get home directory: %v", err)
			}

			if got != tt.want {
				t.Errorf("output mismatch, want: %q, got: %q", tt.want, got)
			}

			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}
