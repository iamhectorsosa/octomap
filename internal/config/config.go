package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

func New(args []string) (*entity.Config, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("Usage: user/repo [--dir] [--branch] [--include] [--exclude] [--output]")
	}

	repo := args[1]
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return nil, fmt.Errorf("error user/repo format: %s", repo)
	}

	fs := flag.NewFlagSet("flags", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", "", "Target directory within the repository")
	branch := fs.String("branch", "main", "Branch to clone (default: main)")
	includeSuffixes := fs.String("include", "", "Comma-separated list of included file extensions")
	excludeSuffixes := fs.String("exclude", "", "Comma-separated list of excluded file extensions")
	output := fs.String("output", "", "Output directory")

	err := fs.Parse(args[2:])
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %v", err)
	}

	include := strings.Split(*includeSuffixes, ",")
	exclude := strings.Split(*excludeSuffixes, ",")

	if len(include) == 1 && include[0] == "" {
		include = nil
	}
	if len(exclude) == 1 && exclude[0] == "" {
		exclude = nil
	}

	repoName := repoParts[1]
	url := fmt.Sprintf("https://github.com/%s/archive/refs/heads/%s.tar.gz", repo, *branch)
	resolvedDir := fmt.Sprintf("%s-%s", repoName, *branch)

	if *dir != "" {
		resolvedDir += "/" + *dir
	}

	resolvedPath, err := resolvePath(*output)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve path: %v", err)
	}

	if resolvedPath != "" {
		ext := filepath.Ext(resolvedPath)
		if ext != "" {
			return nil, fmt.Errorf("invalid directory format: %s", *output)
		}

		_, err = os.Stat(resolvedPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("directory doesn't exist: %v", err)
			}
			return nil, fmt.Errorf("error checking directory: %v", err)
		}
	}

	return &entity.Config{
		Repo:    repoName,
		Dir:     resolvedDir,
		Url:     url,
		Include: include,
		Exclude: exclude,
		Output:  resolvedPath,
	}, nil
}

func resolvePath(path string) (string, error) {
	path = os.ExpandEnv(path)

	if len(path) > 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("reading home dir: %v", err)
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
