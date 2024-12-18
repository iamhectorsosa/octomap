# Octomap 🐙

![Octomap Demo](./demo.gif)

Octomap is a CLI tool that transforms GitHub repositories into structured JSON, making repository content easily digestible for large language models (LLMs) and AI-powered code analysis.

Ideal for developers and data scientists looking to feed repository data directly into AI tools for code understanding, analysis, or transformation.

## Quick Overview

- Download a GitHub repository's contents
- Filter files by branch, directories and/or extensions
- Convert the repository structure into a clean, hierarchical JSON format
- Prepare code repositories for AI-powered processing and analysis

## Requirements

- Go 1.23.2 or later
- Git

## Installation

### Using Go Install

```bash
go install github.com/iamhectorsosa/octomap@latest
```

### From Source

1. Clone the repository:

```bash
git clone https://github.com/iamhectorsosa/octomap.git
cd octomap
```

2. Build the project:

```bash
go build -o octomap
```

3. (Optional) Install the binary:

```bash
go install
```

## Usage

### Basic Usage

```bash
octomap user/repo
```

This command will download the main branch of the specified repository and generate a JSON file with its structure.

### Advanced Options

```bash
# Specify a different branch
octomap user/repo --branch develop

# Target a specific directory within the repository
octomap user/repo --dir src

# Include only specific file types
octomap user/repo --include .go,.proto

# Exclude specific file types
octomap user/repo --exclude .mod,.sum

# Specify a custom output directory
octomap user/repo --output ~/documents
```

### Flags

- `--dir`: Target directory within the repository
- `--branch`: Branch to clone (default: main)
- `--include`: Comma-separated list of included file extensions
- `--exclude`: Comma-separated list of excluded file extensions
- `--output`: Output directory for the generated JSON file

## Development

### Setup

1. Clone the repository
2. Install dependencies:

```bash
go mod download
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build
```
