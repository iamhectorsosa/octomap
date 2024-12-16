package entity

type Config struct {
	Repo    string
	Dir     string
	Url     string
	Output  string
	Include []string
	Exclude []string
}

type Update struct {
	Err         error
	Description string
}
