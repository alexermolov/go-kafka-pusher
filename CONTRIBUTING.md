# Contributing to Kafka Pusher

First off, thank you for considering contributing to Kafka Pusher! It's people like you that make this tool better for everyone.

## Code of Conduct

This project and everyone participating in it is governed by our commitment to providing a welcoming and inspiring community for all.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps to reproduce the problem**
* **Provide specific examples**
* **Describe the behavior you observed and what you expected**
* **Include logs and screenshots if relevant**
* **Specify your environment** (OS, Go version, Kafka version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

* **Use a clear and descriptive title**
* **Provide a detailed description of the proposed functionality**
* **Explain why this enhancement would be useful**
* **List any similar features in other tools**

### Pull Requests

* Fill in the required template
* Follow the Go coding style (use `gofmt` and `golangci-lint`)
* Include tests for new functionality
* Update documentation as needed
* End all files with a newline

## Development Setup

1. Fork and clone the repository
2. Install Go 1.22 or higher
3. Install dependencies: `go mod download`
4. Start local Kafka: `make kafka-up`
5. Run tests: `make test`

## Coding Standards

### Go Code Style

* Follow [Effective Go](https://golang.org/doc/effective_go.html)
* Use `gofmt` for formatting
* Run `golangci-lint` before committing
* Write clear, self-documenting code
* Add comments for exported functions and types

### Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters
* Reference issues and pull requests liberally

Example:
```
Add batch sending support

- Implement batch message accumulation
- Add batch size configuration
- Update documentation

Fixes #123
```

### Testing

* Write unit tests for new features
* Ensure all tests pass: `make test`
* Run race detector: `make test-race`
* Maintain or improve code coverage

### Documentation

* Update README.md for user-facing changes
* Add godoc comments for exported functions
* Update configuration examples if needed
* Include examples for new features

## Project Structure

```
.
├── cmd/kafka-pusher/       # Main application entry point
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── kafka/             # Kafka producer implementation
│   ├── logger/            # Logging setup
│   ├── scheduler/         # Task scheduling
│   └── template/          # Template generation
├── .github/workflows/     # CI/CD pipelines
├── config.example.yaml    # Example configuration
└── payload.example.yaml   # Example payload template
```

## Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create a new tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push the tag: `git push origin v1.0.0`
5. GitHub Actions will handle the release

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
