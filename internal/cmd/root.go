package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iamhectorsosa/octomap/internal/entity"
	"github.com/iamhectorsosa/octomap/internal/model"
	"github.com/spf13/cobra"
)

var (
	repo    string
	branch  string
	url     string
	dir     string
	include []string
	exclude []string
	output  string
)

func init() {
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to clone")
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Target directory within the repository")
	rootCmd.Flags().StringSliceVarP(&include, "include", "i", []string{}, "Comma-separated list of included file extensions")
	rootCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "Comma-separated list of excluded file extensions")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output directory for the generated JSON file")
}

var rootCmd = &cobra.Command{
	Use:   "octomap [user/repo]",
	Short: "Transform GitHub repositories into structured JSON",
	Long:  "Octomap is a CLI tool that transforms GitHub repositories into structured JSON",
	Args: func(cmd *cobra.Command, args []string) error {
		return validateRootArgs(args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Help menu
		if len(args) < 1 {
			return nil
		}

		// GitHub Repository Details
		slug := args[0]
		if err := validateSlug(slug); err != nil {
			return err
		}
		if err := validateBranch(branch); err != nil {
			return err
		}
		repo, url, dir = createRepoDetails(slug, branch, dir)

		// Output Directory
		resolvedOutput, err := resolveOutput(output)
		if err != nil {
			return err
		}
		if err := validateOutput(resolvedOutput); err != nil {
			return err
		}
		output = resolvedOutput
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}

		cfg := &entity.Config{
			RepoName: repo,
			Url:      url,
			Dir:      dir,
			Include:  include,
			Exclude:  exclude,
			Output:   output,
		}

		if _, err := tea.NewProgram(model.New(cfg)).Run(); err != nil {
			return err
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
