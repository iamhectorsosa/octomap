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

func main() {
	// TODO: Abstract this into a function
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . user/repo [--dir DIR] [--branch BRANCH] [--include .go,.proto] [--exclude .mod,.sum]")
		os.Exit(1)
	}

	repo := os.Args[1]

	fs := flag.NewFlagSet("flags", flag.ContinueOnError)

	dir := fs.String("dir", "", "Target directory within the repository")
	branch := fs.String("branch", "main", "Branch to clone (default: main)")
	includeSuffixes := fs.String("include", "", "Comma-separated list of included file extensions")
	excludeSuffixes := fs.String("exclude", "", "Comma-separated list of excluded file extensions")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	include := strings.Split(*includeSuffixes, ",")
	exclude := strings.Split(*excludeSuffixes, ",")

	if len(include) == 1 && include[0] == "" {
		include = nil
	}
	if len(exclude) == 1 && exclude[0] == "" {
		exclude = nil
	}

	if _, err := tea.NewProgram(newModel(repo, *dir, *branch, include, exclude)).Run(); err != nil {
		fmt.Println("Error starting Bubble Tea program:", err)
		os.Exit(1)
	}
}

type config struct {
	repo    string
	dir     string
	branch  string
	include []string
	exclude []string
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

func newModel(repo, dir, branch string, include, exclude []string) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	config := config{
		repo:    repo,
		dir:     dir,
		branch:  branch,
		include: include,
		exclude: exclude,
	}

	return model{
		config:  config,
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
			go checkRepo(m.config.repo, m.config.dir, m.config.branch, m.config.include, m.config.exclude, m.ch, 25*time.Millisecond)
		}

		msg, ok := <-m.ch
		if !ok {
			return processEndMsg{}
		}
		return processFinishedMsg(msg.(string))
	}
}

// TODO: Add outputDir as a flag
const (
	dataOutputFile   = ""
	githubTarballURL = "https://github.com/%s/archive/refs/heads/%s.tar.gz"
)

// TODO: Rename function
func checkRepo(inputRepo, inputTargetDir, inputBranch string, include, exclude []string, ch chan<- interface{}, delay time.Duration) {
	defer close(ch)

	// TODO: Create a validate arguments function
	repoParts := strings.Split(inputRepo, "/")
	if len(repoParts) != 2 {
		log.Fatalf("Invalid repository format. Expected 'user/repo', got: %q", inputRepo)
	}

	repo := repoParts[1]
	tarballURL := fmt.Sprintf(githubTarballURL, inputRepo, inputBranch)
	targetDir := fmt.Sprintf("%s-%s", repo, inputBranch)

	if inputTargetDir != "" {
		targetDir += "/" + inputTargetDir
	}

	resp, err := http.Get(tarballURL)
	if err != nil {
		log.Fatalf("Error fetching tarball: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch tarball: %s", resp.Status)
	}

	data, err := tarballReader(targetDir, include, exclude, resp.Body, ch, delay)
	if err != nil {
		log.Fatalf("error=%v", err)
	}

	// TODO: Create encoder function
	if dataOutputFile != "" {
		ext := filepath.Ext(dataOutputFile)
		if ext != "" {
			log.Fatalf("Has to be a valid path without extensions")
		}

		_, err = os.Stat(dataOutputFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("directory does not exist: %v", dataOutputFile)
			}
			log.Fatalf("error checking directory: %v", err)
		}
	}

	currentDateTime := time.Now().Format("20060102_150405")
	newFileName := fmt.Sprintf("%s%s%s.json", dataOutputFile, repo, currentDateTime)

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
