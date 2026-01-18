# Contributing to GoGitSomePrivacy

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- GitHub account

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/GoGitSomePrivacy.git
   cd GoGitSomePrivacy
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/h4n0sh1/GoGitSomePrivacy.git
   ```

4. Install dependencies:
   ```bash
   make deps
   ```

5. Run tests to ensure everything works:
   ```bash
   make test
   ```

## Development Workflow

### Creating a Feature Branch

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### Making Changes

1. Write your code following the project's style guidelines
2. Add tests for new functionality
3. Update documentation as needed
4. Run tests and linter:
   ```bash
   make test
   make lint
   make fmt
   ```

### Committing Changes

Follow conventional commit format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Example:
```bash
git commit -m "feat(scanner): add support for email detection"
```

### Pushing Changes

```bash
git push origin feature/your-feature-name
```

### Creating a Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select your feature branch
4. Fill in the PR template with:
   - Description of changes
   - Related issue numbers
   - Testing performed
   - Screenshots (if applicable)

## Code Style Guidelines

### Go Style

Follow standard Go conventions:

- Use `gofmt` for formatting (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names
- Keep functions focused and small
- Document exported functions and types

### Package Organization

```
internal/  - Private application code
pkg/       - Public library code
cmd/       - Command-line tools
```

### Naming Conventions

- **Packages**: Short, lowercase, single-word names
- **Files**: Lowercase with underscores (`word_word.go`)
- **Functions**: CamelCase, exported functions start with uppercase
- **Variables**: camelCase for local, CamelCase for exported
- **Constants**: CamelCase or SCREAMING_SNAKE_CASE for groups

### Comments

```go
// Package scanner provides functionality for scanning GitHub commits.
package scanner

// Scanner orchestrates the scanning process for GitHub commits.
// It coordinates between the GitHub client, PII detector, and worker pool.
type Scanner struct {
    // ...
}

// NewScanner creates a new scanner instance with the given configuration.
// It initializes the detector and worker pool with appropriate defaults.
func NewScanner(client *github.Client, criteria models.PIISearchCriteria, config Config) *Scanner {
    // ...
}
```

## Testing Guidelines

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Use meaningful test names

Example:
```go
func TestDetector_DetectInText(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantMatches int
    }{
        {
            name:        "finds full name",
            input:       "John Doe committed this",
            wantMatches: 1,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/scanner/...

# Run with coverage
make coverage

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Test Coverage

- Aim for >80% coverage for new code
- Critical paths should have 100% coverage
- Don't sacrifice quality for coverage numbers

## Documentation

### Code Documentation

- Document all exported types, functions, and constants
- Include examples for complex functionality
- Keep comments up to date with code changes

### User Documentation

Update relevant documentation files:
- `README.md`: Overview and quick start
- `docs/USAGE.md`: Detailed usage instructions
- `docs/ARCHITECTURE.md`: Architecture decisions

## Pull Request Guidelines

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] Tests pass locally (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### PR Description Template

```markdown
## Description
Brief description of changes

## Related Issues
Fixes #123

## Changes Made
- Added feature X
- Fixed bug Y
- Updated documentation Z

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing performed
- [ ] Edge cases considered

## Checklist
- [ ] Tests pass
- [ ] Linter passes
- [ ] Documentation updated
- [ ] Breaking changes documented
```

## Review Process

1. Automated checks run on PR creation
2. Maintainer reviews code
3. Address feedback with new commits
4. Once approved, maintainer will merge

### Addressing Feedback

- Push new commits (don't force push)
- Respond to comments
- Mark conversations as resolved when addressed

## Release Process

Maintainers handle releases following semantic versioning:

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes

## Issue Guidelines

### Reporting Bugs

Include:
- Go version
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs/output

### Suggesting Features

Include:
- Use case description
- Proposed solution
- Alternative solutions considered
- Potential impacts

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Documentation improvements
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention needed

## Community

### Getting Help

- Open an issue for bugs or features
- Start a discussion for questions
- Check existing issues before opening new ones

### Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to GoGitSomePrivacy!
