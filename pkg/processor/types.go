package processor

type RepositoryData map[string]interface{}

type Config struct {
	Repo    string
	Url     string
	Dir     string
	Output  string
	Include []string
	Exclude []string
	Stdout  bool
}

type Update struct {
	Err         error
	Description string
}

type Processor struct {
	config        *Config
	data          RepositoryData
	ch            chan<- Update
	dirCount      int
	fileCount     int
	dataFileCount int
}
