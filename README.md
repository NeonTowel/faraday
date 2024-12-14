# Faraday

Faraday is a command-line Go application designed to interact with an AI service. It takes user prompts as input and communicates with the service to provide responses. This project is configured using a YAML file and supports building for multiple operating systems and architectures.

## Features

- Command-line interface for easy interaction.
- Supports configuration via a YAML file.
- Utilizes spinner for user-friendly loading animations.
- Supports multiple OS and architecture builds.

## Installation

### Prerequisites

- Go 1.23 or later
- [Task](https://taskfile.dev/) for task management

### Building the Project

To build the project, run:

```sh
task build
```

### Building for Release

To build the project for release across different platforms, run:

```sh
task release
```

## Configuration

The application requires a `config.yaml` file located in the same directory as the executable. The configuration file should contain the following structure:

```yaml
api:
  url: "https://api.example.com"
  key: "your-api-key"
```

## Usage

Run the application with a prompt:

```sh
faraday "Your prompt here"
```

To include a context file, use the `@file` syntax:

```sh
faraday "Your prompt here" @path/to/context.file
```

## Dependencies

The project relies on several Go packages, including:

- `github.com/briandowns/spinner` for loading animations.
- `github.com/charmbracelet/glamour` for rendering styled text.
- `gopkg.in/yaml.v3` for YAML parsing.

For a complete list of dependencies, refer to the `go.mod` file.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
