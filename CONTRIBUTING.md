# Contributing to inceptools

Thank you for your interest in contributing to **inceptools**! This guide will help you get started.

## 🐛 Reporting Issues

- Use the [GitHub Issues](https://github.com/IncepTools/inceptools-cli/issues) page
- Include your Go version, OS, database dialect, and steps to reproduce
- Attach relevant logs or error output

## 🛠️ Development Setup

```bash
# Clone the repository
git clone https://github.com/IncepTools/inceptools-cli.git
cd inceptools-cli

# Install dependencies
go mod download

# Build
go build -o inceptools .

# Run vet
go vet ./...
```

## 📝 Submitting Changes

1. **Fork** the repository
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
3. **Make your changes** — keep commits focused and atomic
4. **Run checks** before pushing:
   ```bash
   go vet ./...
   go build ./...
   ```
5. **Open a Pull Request** against `main` with a clear description

## 💬 Commit Convention

Use clear, descriptive commit messages:

```
feat: add support for CockroachDB dialect
fix: handle empty migration directory gracefully
docs: update installation instructions
refactor: simplify DSN resolution logic
```

## 📏 Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and well-documented
- Use meaningful variable and function names

## 📄 License

By contributing, you agree that your contributions will be licensed under the [GPL-3.0 License](LICENSE).
