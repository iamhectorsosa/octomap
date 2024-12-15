package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	mainStyle = lipgloss.NewStyle().MarginLeft(1)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

type config struct {
	repo    string
	dir     string
	branch  string
	url     string
	output  string
	include []string
	exclude []string
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

func getConfig() (cfg *config, err error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("Usage: user/repo [--dir] [--branch] [--include] [--exclude] [--output]")
	}

	repo := os.Args[1]
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

	err = fs.Parse(os.Args[2:])
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

	return &config{
		repo:    repoName,
		dir:     resolvedDir,
		url:     url,
		branch:  *branch,
		include: include,
		exclude: exclude,
		output:  resolvedPath,
	}, nil
}

func main() {
	cfg, err := getConfig()
	if err != nil {
		fmt.Println("Error handling arguments")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(cfg)).Run(); err != nil {
		fmt.Println("Error starting Bubble Tea program:", err)
		os.Exit(1)
	}
}

type result struct {
	directory string
}

type model struct {
	config   config
	ch       chan interface{}
	results  []result
	spinner  spinner.Model
	complete bool
	quitting bool
}

func newModel(cfg *config) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		config:  *cfg,
		spinner: sp,
		results: []result{},
		ch:      make(chan interface{}),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runPretendProcess(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case processFinishedMsg:
		res := result{directory: string(msg)}
		m.results = append(m.results, res)
		if len(m.results) > 10 {
			m.results = m.results[1:]
		}
		return m, m.runPretendProcess()
	case processEndMsg:
		m.complete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	s := "\n"
	if !m.complete {
		s += m.spinner.View()
	} else {
		s += "  "
	}
	s += "🐙 Mapping repository...\n\n"

	for _, res := range m.results {
		s += fmt.Sprintf("%s %s\n", checkMark, res.directory)
	}

	if !m.complete {
		s += helpStyle("\nPress any key to exit")
	} else {
		s += "\nMapping complete!\n\n"
	}

	return mainStyle.Render(s)
}

type (
	processEndMsg      struct{}
	processFinishedMsg string
)

func (m model) runPretendProcess() tea.Cmd {
	return func() tea.Msg {
		if len(m.results) == 0 {
			go checkRepo(m.config.repo, m.config.url, m.config.dir, m.config.output, m.config.include, m.config.exclude, m.ch, 25*time.Millisecond)
		}

		msg, ok := <-m.ch
		if !ok {
			return processEndMsg{}
		}
		return processFinishedMsg(msg.(string))
	}
}

// TODO: Rename function
func checkRepo(repo, url, dir, output string, include, exclude []string, ch chan<- interface{}, delay time.Duration) {
	defer close(ch)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching tarball: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch tarball: %s", resp.Status)
	}

	data, err := tarballReader(dir, include, exclude, resp.Body, ch, delay)
	if err != nil {
		log.Fatalf("error=%v", err)
	}

	// TODO: Create encoder function
	currentDateTime := time.Now().Format("20060102_150405")
	newFileName := fmt.Sprintf("%s%s%s.json", output, repo, currentDateTime)

	f, err := os.Create(newFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}

func tarballReader(dir string, include, exclude []string, r io.Reader, ch chan<- interface{}, delay time.Duration) (map[string]interface{}, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("error decompressing tarbal: %v", err)
	}
	defer gzipReader.Close()

	data := make(map[string]interface{})

	tarReader := tar.NewReader(gzipReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tarball: %v", err)
		}

		if hdr.Typeflag == tar.TypeDir || !strings.HasPrefix(hdr.Name, dir) {
			continue
		}

		relativePath := strings.TrimPrefix(hdr.Name, dir+"/")

		shouldProcess := true
		if len(include) > 0 {
			shouldProcess = false
			for _, suffix := range include {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = true
					break
				}
			}
		}
		if shouldProcess && len(exclude) > 0 {
			for _, suffix := range exclude {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = false
					break
				}
			}
		}

		if shouldProcess {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				return nil, fmt.Errorf("error reading file: %s - %v", hdr.Name, err)
			}

			pathParts := strings.Split(relativePath, "/")

			current := data
			for i, part := range pathParts {
				if i == len(pathParts)-1 {
					current[part] = buf.String()
					time.Sleep(delay)
					ch <- relativePath
				} else {
					if _, exists := current[part]; !exists {
						current[part] = make(map[string]interface{})
					}

					var ok bool
					current, ok = current[part].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("unexpected structure found on: %s", hdr.Name)
					}
				}
			}
		}
	}

	return data, nil
}
