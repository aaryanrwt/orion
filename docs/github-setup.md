# GitHub Repository Settings & Branch Protection Guide

This guide details the recommended GitHub repository configuration for **Orion**. These settings are designed to enforce security, ensure release quality, and maintain code consistency for both solo maintainers and growing contributor communities.

---

## 1. Branch Protection Rules

Branch protection rules prevent accidental deletions, force pushes, and ensure code meets linting/vetting guidelines before being merged.

### Target: `main`

To set up protection rules:
1. Navigate to your repository on GitHub.
2. Click **Settings** -> **Branches**.
3. Under **Branch protection rules**, click **Add branch protection rule**.
4. Set **Branch name pattern** to `main`.

### Recommended Configurations

*   **Require a pull request before merging**: `Enabled`
    *   *Solo Maintainer Note*: Check **Allow bypasses for pull requests** and select your administrator account to allow direct merges when necessary without a PR review round.
*   **Require status checks to pass before merging**: `Enabled`
    *   Verify that matrix builds and lints succeed before code is merged.
    *   **Require branches to be up to date before merging**: `Enabled` (prevents merging stale code that might conflict with recent updates).
    *   **Status Checks to Require**:
        *   `Lint & Format Check` (handles formatting, mod tidy, and `go vet`).
        *   `Build & Test (ubuntu-latest)`
        *   `Build & Test (macos-latest)`
        *   `Build & Test (windows-latest)`
        *   `CodeQL Analysis`
*   **Require conversation resolution before merging**: `Enabled`
    *   Ensures all reviewer threads and developer questions are resolved before code is integrated.
*   **Restrict who can push to matching branches**: `Disabled`
    *   Keep this disabled for solo maintainer setups to prevent locking yourself out, but enable it once multiple administrators or deploy keys are configured.
*   **Block force pushes**: `Enabled`
    *   Guarantees the integrity of the commit history.
*   **Require signed commits**: `Enabled` (Optional but highly recommended)
    *   Ensures that commits can be cryptographically traced to verified developer accounts.

---

## 2. Repository Security Settings

Enabling advanced GitHub security features builds trust for down-stream consumers:

1.  **Dependency Graph & Dependabot**:
    *   Navigate to **Settings** -> **Code security and analysis**.
    *   Enable **Dependency graph**, **Dependabot alerts**, and **Dependabot security updates**.
    *   This pairs with our configured `.github/dependabot.yml` to automatically submit PRs updating vulnerable packages.
2.  **CodeQL Vulnerability Scanning**:
    *   Enables automated security scans on every pull request. This runs using `.github/workflows/codeql.yml`.

---

## 3. General Settings & Actions Permissions

Under **Settings** -> **General**:
*   **Pull Requests**:
    *   Select **Allow squash merging** and uncheck other merge options. This keeps the commit history of the `main` branch linear and clean.
    *   Enable **Automatically delete head branches** to clean up feature branches once a PR is merged.
