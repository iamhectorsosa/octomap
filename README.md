# Octomap 🐙

![Octomap Demo](./demo.gif)

Octomap is a CLI tool that transforms GitHub repositories into structured JSON, making repository content easily digestible for large language models (LLMs) and AI-powered code analysis.

## Quick Overview

With Octomap, you can:

- Download a GitHub repository's contents
- Filter files by branch, directories and/or extensions
- Convert the repository structure into a clean, hierarchical JSON format
- Prepare code repositories for AI-powered processing and analysis

Ideal for developers and data scientists looking to feed repository data directly into AI tools for code understanding, analysis, or transformation.

Simply specify a GitHub repo, and Octomap does the heavy lifting of extracting and structuring its contents.

**How does it work?** Octomap downloads a repository's tarball, processes the files based on specified filters, and recursively builds a nested JSON map representing the repository's directory structure and file contents.
