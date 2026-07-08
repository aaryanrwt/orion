# Contributing to Orion

We are excited that you are interested in contributing to Orion! As a contributor, you help make Orion the best peer-to-peer developer compute mesh.

---

## 1. Development Environment Setup

### Prerequisites
- Go **1.22** or higher.
- A functional C/C++ compiler (gcc/clang) if you want to run CGO-enabled tests (such as standard `-race` test suites).

### Cloning & Compilation
```bash
git clone https://github.com/orion-infra/orion.git
cd orion
go build -o orion ./cmd/orion
```

---

## 2. Coding Standards

We follow standard Go practices:
*   **Format**: Run `go fmt ./...` before submitting your PR.
*   **Linting**: Run `go vet ./...` to check for static issues.
*   **Style**: Avoid external dependencies where possible. Ensure all comments and documentation are clear and grammatically correct.

---

## 3. Contribution Lifecycle

### Step 1: Fork and Branch
1. Fork the repository on GitHub.
2. Clone your fork locally.
3. Create a branch from `main` with a descriptive name:
   ```bash
   git checkout -b feature/dynamic-allocation
   ```

### Step 2: Make Changes & Test
1. Make your code modifications.
2. Add corresponding tests for your changes.
3. Run the complete test suite locally to verify:
   ```bash
   go test -count=1 ./...
   ```
4. Verify code cleanliness:
   ```bash
   go vet ./...
   ```

### Step 3: Commit and Push
*   Write clear, semantic commit messages (e.g. `feat: implement active transport matching`, `fix: correct memory allocation boundary`).
*   Push your branch to your GitHub fork.

### Step 4: Open a Pull Request
1. Submit your pull request to the `main` branch of `orion-infra/orion`.
2. Fill out the PR template completely.
3. Keep PRs focused. If you are proposing multiple changes, submit them as separate pull requests.

---

## 4. Review & Merge Process
1. **Automated Testing**: Your PR will trigger CI workflows verifying formats, tidiness, vet compliance, and cross-platform tests.
2. **Review Check**: At least one core maintainer must review and approve your pull request.
3. **Squash & Merge**: Once approved and all checks pass, we will squash and merge your changes into `main`.
