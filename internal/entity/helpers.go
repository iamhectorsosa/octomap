package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	GITHUB_ARCHIVE = "https://github.com/%s/%s/archive/refs/heads/%s.tar.gz"

	invalidUserRepoTxt   = "invalid [user/repo] input, received %q\n"
	invalidBranchName    = "invalid branch, received %q\n"
	invalidOutputWithExt = "invalid output, cannot contain extension, received %q\n"

	errHomeDirectory    = "failed to get user home directory, %v\n"
	errOutputDoesntExit = "output path does not exist, received %q\n%v\n"
	errOutputAccess     = "output path cannot be accessed, received %q\n%v\n"
)

func validateSlug(slug string) error {
	validatedSlug := strings.Split(slug, "/")

	if len(validatedSlug) != 2 {
		return fmt.Errorf(invalidUserRepoTxt, slug)
	}

	if len(validatedSlug[0]) == 0 || len(validatedSlug[1]) == 0 {
		return fmt.Errorf(invalidUserRepoTxt, slug)
	}
	return nil
}

func validateBranch(branch string) error {
	if len(branch) == 0 {
		return fmt.Errorf(invalidBranchName, branch)
	}
	return nil
}

func createRepoDetails(slug string, branch, inputDir string) (repo, url, dir string) {
	validatedSlug := strings.SplitN(slug, "/", 2)
	user := validatedSlug[0]
	repo = validatedSlug[1]

	url = fmt.Sprintf(GITHUB_ARCHIVE, user, repo, branch)

	dir = fmt.Sprintf("%s-%s", repo, branch)
	if inputDir != "" {
		dir += "/" + inputDir
	}

	return
}

func validateOutput(output string) error {
	if output == "" {
		return nil
	}

	ext := filepath.Ext(output)
	if ext != "" {
		return fmt.Errorf(invalidOutputWithExt, output)
	}

	_, err := os.Stat(output)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(errOutputDoesntExit, output, err)
		}
		return fmt.Errorf(errOutputAccess, output, err)
	}

	return nil
}

func resolveOutput(output string) (string, error) {
	if output == "" {
		return output, nil
	}

	expandedOutput := os.ExpandEnv(output)

	if !strings.HasPrefix(expandedOutput, "~/") {
		return expandedOutput, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf(errHomeDirectory, err)
	}

	return filepath.Join(home, expandedOutput[:2]), nil
}
