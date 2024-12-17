package entity

type Config struct {
	RepoName string
	Dir      string
	Url      string
	Output   string
	Include  []string
	Exclude  []string
}

type Update struct {
	Err         error
	Description string
}
