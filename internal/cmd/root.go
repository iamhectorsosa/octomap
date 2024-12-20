package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iamhectorsosa/octomap/internal/model"
	"github.com/iamhectorsosa/octomap/pkg/processor"
	"github.com/spf13/cobra"
)

var (
	branch  string
	dir     string
	include []string
	exclude []string
	output  string
	stdout  bool
)

func init() {
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to clone")
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Target directory within the repository")
	rootCmd.Flags().StringSliceVarP(&include, "include", "i", []string{}, "Comma-separated list of included file extensions")
	rootCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "Comma-separated list of excluded file extensions")
	rootCmd.Flags().BoolVarP(&stdout, "stdout", "s", false, "Output to stdout. Note: output will be ignored.")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output directory for the generated JSON file")
}

var rootCmd = &cobra.Command{
	Use:   "octomap [user/repo]",
	Short: "Transform GitHub repositories into structured JSON",
	Long:  "Octomap is a CLI tool that transforms GitHub repositories into structured JSON",
	Args: func(cmd *cobra.Command, args []string) error {
		return validateRootArgs(args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Help()
			return nil
		}

		slug := args[0]
		config, err := processor.NewConfig(slug, branch, dir, output, stdout, include, exclude)
		if err != nil {
			return err
		}

		// If stdout wasn't provided run the Bubbletea program
		if !stdout {
			if _, err := tea.NewProgram(model.New(config)).Run(); err != nil {
				return err
			}
			return nil
		}

		// If stdout was provided, create a new processor and run a process
		p := processor.New(config, nil)
		data, err := p.Process(0)
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, data)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
