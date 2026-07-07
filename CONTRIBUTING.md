# Contributing to Orion

We are excited that you want to contribute to Orion! We want to keep the development process clean, focused, and enjoyable.

Before contributing, please read this document and review our [Philosophy](docs/philosophy.md) to align on our product principles.

---

## Code of Conduct

We aim to foster an open, welcoming, and professional community. Treat all contributors and users with respect.

## Design Philosophy Check

Every feature or change must align with our core design goals:
1. **Zero protocol details** exposed to the user.
2. **Helpful errors** detailing *What happened*, *Why*, and *What to Try*.
3. **No performance regressions** (startup time must remain sub-50ms).
4. **No telemetry** by default.

If a proposed feature adds complexity without improving the "v0.1 core user journey" (Install → Init → Join → Run), it should not be merged.

## Local Development Setup

To work on Orion, you will need:
* **Go 1.22** or later installed.
* Standard development tools (`git`, `make` if applicable).

### Running Tests

We enforce unit tests and integration tests for every command:

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
```

### Formatting and Linting

We run strict formatting and linting checks on every commit:

```bash
# Format code
go fmt ./...

# Run golangci-lint (install via: https://golangci-lint.run/)
golangci-lint run
```

---

## Pull Request Guidelines

1. **Keep it focused**: One PR should do one thing. If you want to make multiple unrelated improvements, please open separate PRs.
2. **Write tests**: Every new CLI command, flag, or business logic change must include unit or integration tests.
3. **Document changes**: Update help strings, CLI examples, or documentation files (like `README.md`) if your changes introduce user-facing modifications.
4. **Clean Git History**: Commit messages should be clear, concise, and explain the *why* of the change.

---

## Contact & Issues

* Search existing issues before creating a new one.
* When opening a bug report, **always run `orion doctor`** and attach the terminal report.
