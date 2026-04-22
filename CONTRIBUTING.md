# Contributing to kmgr

Thanks for your interest in contributing to kmgr! This document provides guidelines and instructions for contributing.

---

## Development Setup

### Prerequisites

- **Go 1.22+**
- **Git**

### Getting Started

```bash
# Clone the repository
git clone https://github.com/UnPoilTefal/kmgr.git
cd kmgr

# Install dependencies
go mod download

# Build the binary
make build

# Run tests
make test
```

---

## Workflow

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create a branch** for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** and ensure tests pass:
   ```bash
   make check  # Run tests and linters
   ```
5. **Commit** with clear, descriptive messages:
   ```bash
   git commit -m "Add feature: description"
   ```
6. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request** on GitHub

---

## Code Standards

### Go Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (applied automatically by `make fmt`)
- Write clear, idiomatic Go
- Add comments for exported functions and packages

### Testing

- Write unit tests for new functionality
- Tests should be in `*_test.go` files alongside the code they test
- Aim for reasonable coverage (70%+ for critical paths)

```bash
# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...
```

### Naming Convention

kmgr enforces a strict naming convention for kubeconfigs:

```
File format  : kubeconfig_{user}@{cluster}.yaml
Context name : {user}@{cluster}
```

Any code changes must respect and maintain this convention.

---

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add feature X
fix: resolve issue with Y
docs: update README
test: add tests for Z
chores: update dependencies
refactor: improve code structure
```

---

## Pull Request Guidelines

- One feature/fix per PR
- Include tests for new functionality
- Update README if user-facing changes
- Reference issues: "Closes #123"
- Keep commits clean and squash if needed

---

## Project Structure

```
kmgr/
├── cmd/              # CLI commands (cobra)
├── internal/
│   ├── config/       # kubeconfig management (k8s.io/client-go)
│   └── normalize/    # naming normalization
├── docs/             # Documentation and reference implementations
├── main.go           # Entry point
├── go.mod, go.sum    # Dependencies
├── Makefile          # Build and test targets
└── README.md         # User documentation
```

---

## Key Patterns

### Adding a New Command

1. Create `cmd/yourcommand.go`:
   ```go
   package cmd

   import "github.com/spf13/cobra"

   var yourCmd = &cobra.Command{
       Use:   "yourcommand",
       Short: "Description",
       RunE: runYourCommand,
   }

   func init() {
       rootCmd.AddCommand(yourCmd)
   }

   func runYourCommand(cmd *cobra.Command, args []string) error {
       // Implementation
       return nil
   }
   ```

2. Import it in `cmd/root.go`:
   ```go
   func init() {
       rootCmd.AddCommand(yourCmd)
   }
   ```

3. Add tests in `cmd/yourcommand_test.go`

### Configuration Management

Use the `internal/config` package for kubeconfig operations. Always:
- Use `k8s.io/client-go` for parsing/writing kubeconfigs
- Respect the naming convention
- Set file permissions to 0600
- Create backups before modifications

---

## Testing

### Unit Tests

```bash
go test ./...
go test -v ./cmd  # Verbose output
go test -race ./... # With race detector
```

### Manual Testing

```bash
make build
./bin/kmgr init
./bin/kmgr import -f test_kubeconfig.yaml -u testuser -c testcluster
./bin/kmgr list
```

---

## Documentation

- Update [README.md](README.md) for user-facing changes
- Add comments for complex logic
- Reference [docs/reference.sh](docs/reference.sh) for behavioral guidance
- Keep CLAUDE.md as internal development notes (not published)

---

## Reporting Issues

When reporting bugs or requesting features, include:

- **Current behavior**: What happens now?
- **Expected behavior**: What should happen?
- **Steps to reproduce**: How do we repeat the issue?
- **Environment**: OS, Go version, kmgr version
- **Example kubeconfigs** (sanitized if needed)

---

## Code Review

All PRs require review before merging. Reviewers will check for:

- Adherence to naming conventions
- Test coverage
- Documentation
- Performance implications
- Security issues

---

## License

By contributing to kmgr, you agree that your contributions will be licensed under the MIT License.

---

## Questions?

Open an issue or start a discussion on GitHub. We're happy to help!
