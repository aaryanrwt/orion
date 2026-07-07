# Orion Release Quality Checklist

This checklist defines the quality standards that Orion must pass before any public release tag is pushed.

---

## 1. Automated Quality Gates

- [ ] **Formatting**: Ensure all code is cleanly formatted:
  ```bash
  go fmt ./...
  ```
- [ ] **Linting**: Ensure there are no static analysis warnings:
  ```bash
  golangci-lint run
  ```
- [ ] **Unit & Integration Tests**: Verify all tests pass cleanly:
  ```bash
  go test -v -race ./...
  ```

---

## 2. CLI UX Verification

- [ ] **Command Help Alignment**: Verify every subcommand prints a polished description and exactly one example when `-h` or `--help` is queried.
- [ ] **No Telemetry**: Double check that no external network calls are performed on CLI startup.
- [ ] **NO_COLOR Support**: Validate that `NO_COLOR=1` strips all ANSI escape characters from the output.
- [ ] **Diagnostics**: Run `orion doctor` to verify it provides actionable recovery instructions.
- [ ] **Structured Errors**: Confirm that all failure conditions display a diagnostic error in the *What*, *Why*, *Try* format and avoid printing raw Go stack traces.

---

## 3. Platform Verification

Compile and run verification tests on:
- [ ] **Linux** (amd64 / arm64)
- [ ] **macOS** (amd64 / Apple Silicon)
- [ ] **Windows** (amd64)

---

## 4. Documentation & Release Metadata

- [ ] **README Verification**: Verify all code snippets and examples shown in the README are functional on a fresh environment.
- [ ] **Changelog**: Ensure `CHANGELOG.md` is updated with all features, fixes, and dependencies added in the release tag.
- [ ] **Roadmap**: Sync upcoming features in `ROADMAP.md` with the current milestones.
