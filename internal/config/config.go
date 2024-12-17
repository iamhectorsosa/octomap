package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iamhectorsosa/octomap/internal/entity"
)

var errorHrd = lipgloss.NewStyle().
	SetString("ERROR").
	Bold(true).
	Foreground(lipgloss.Color("204"))

var hintHrd = lipgloss.NewStyle().
	SetString("HINT").
	Bold(true)

const (
	missingArguments     = "missing arguments"
	missingArgumentsHint = "user/repo [--dir] [--branch] [--include] [--exclude] [--output]"
	invalidUserRepo      = "invalid user/repo"
	invalidUserRepoHint  = "repositories must follow the 'user/repo' pattern"
	errParsingFlags      = "cannot not parse flags"
	errResolvingPath     = "cannot resolve path"
	invalidOutputWithExt = "cannot use output with file extension"
	errDirDoesntExist    = "directory doesn't exist"
	errCannotValidateDir = "cannot validate directory"
)

func New(args []string) (*entity.Config, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%s %s\n%s %s", errorHrd, missingArguments, hintHrd, missingArgumentsHint)
	}

	repo := args[1]
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return nil, fmt.Errorf("%s %s\n%s %s", errorHrd, invalidUserRepo, hintHrd, invalidUserRepoHint)
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
		return nil, fmt.Errorf("%s %s\n%v", errorHrd, errParsingFlags, err)
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

	resolvedOutput, err := resolvePath(*output)
	if err != nil {
		return nil, fmt.Errorf("%s %s\n%v", errorHrd, errResolvingPath, err)
	}

	if resolvedOutput != "" {
		ext := filepath.Ext(resolvedOutput)
		if ext != "" {
			return nil, fmt.Errorf("%s %s: %s\n", errorHrd, invalidOutputWithExt, *output)
		}

		_, err = os.Stat(resolvedOutput)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%s %s\n%v", errorHrd, errDirDoesntExist, err)
			}

			return nil, fmt.Errorf("%s %s\n%v", errorHrd, errCannotValidateDir, err)
		}
	}

	return &entity.Config{
		RepoName: repoName,
		Dir:      resolvedDir,
		Url:      url,
		Include:  include,
		Exclude:  exclude,
		Output:   resolvedOutput,
	}, nil
}

func resolvePath(path string) (string, error) {
	path = os.ExpandEnv(path)

	if len(path) > 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
