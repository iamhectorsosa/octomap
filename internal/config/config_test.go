package config_test

import (
	"testing"

	"github.com/iamhectorsosa/octomap/internal/config"
	"github.com/iamhectorsosa/octomap/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		expectedConfig *entity.Config
		name           string
		args           []string
		expectedErr    bool
	}{
		{
			name: "No optional flags, default branch",
			args: []string{"cmd", "user/repo"},
			expectedConfig: &entity.Config{
				RepoName: "repo",
				Dir:      "repo-main",
				Url:      "https://github.com/user/repo/archive/refs/heads/main.tar.gz",
				Include:  nil,
				Exclude:  nil,
				Output:   "",
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
			expectedConfig: &entity.Config{
				RepoName: "repo",
				Dir:      "repo-develop/src",
				Url:      "https://github.com/user/repo/archive/refs/heads/develop.tar.gz",
				Include:  []string{".go", ".proto"},
				Exclude:  []string{".mod", ".sum"},
				Output:   "/Users/hectorsosa/documents",
			},
			expectedErr: false,
		},
		{
			name: "Empty include and exclude flags",
			args: []string{"cmd", "user/repo", "--include", "", "--exclude", ""},
			expectedConfig: &entity.Config{
				RepoName: "repo",
				Dir:      "repo-main",
				Url:      "https://github.com/user/repo/archive/refs/heads/main.tar.gz",
				Include:  nil,
				Exclude:  nil,
				Output:   "",
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
			cfg, err := config.New(tt.args)

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
