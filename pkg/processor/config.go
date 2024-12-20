package processor

func NewConfig(slug, branch, dir, output string, stdout bool, include, exclude []string) (*Config, error) {
	// GitHub Repository Details
	if err := validateSlug(slug); err != nil {
		return nil, err
	}
	if err := validateBranch(branch); err != nil {
		return nil, err
	}
	repo, url, createdDir := createRepoDetails(slug, branch, dir)

	var resolvedOutput string

	// Output Directory
	if !stdout {
		resolvedOutput, err := resolveOutput(output)
		if err != nil {
			return nil, err
		}
		if err := validateOutput(resolvedOutput); err != nil {
			return nil, err
		}
	}

	return &Config{
		Repo:    repo,
		Url:     url,
		Dir:     createdDir,
		Output:  resolvedOutput,
		Stdout:  stdout,
		Include: include,
		Exclude: exclude,
	}, nil
}
