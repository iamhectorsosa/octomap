package entity

type Config struct {
	Repo    string
	Url     string
	Dir     string
	Output  string
	Include []string
	Exclude []string
}

type Update struct {
	Err         error
	Description string
}

func NewConfig(slug, branch, dir, output string, include, exclude []string) (*Config, error) {
	// GitHub Repository Details
	if err := validateSlug(slug); err != nil {
		return nil, err
	}
	if err := validateBranch(branch); err != nil {
		return nil, err
	}
	repo, url, createdDir := createRepoDetails(slug, branch, dir)

	// Output Directory
	resolvedOutput, err := resolveOutput(output)
	if err != nil {
		return nil, err
	}
	if err := validateOutput(resolvedOutput); err != nil {
		return nil, err
	}

	return &Config{
		Repo:    repo,
		Url:     url,
		Dir:     createdDir,
		Output:  resolvedOutput,
		Include: include,
		Exclude: exclude,
	}, nil
}
