# Changelog

All notable changes to the Orion project will be documented in this file.

---

## [0.1.0-beta] - 2026-07-07

Initial release of the Orion CLI developer utility.

### Added
- **Cobra Command Router**: Streamlined 8 core subcommands (`init`, `join`, `devices`, `run`, `status`, `doctor`, `version`, `help`) with sub-50ms CLI startup.
- **Engine Abstraction Layer**: Isolated command runner and device mesh manager behind a unified Go interface (`Engine`), separating CLI logic from distributed networking primitives.
- **High-Fidelity Mock Engine**: Simulates network latencies, node handshakes, device status checks, and streams stdout/stderr lines concurrently.
- **Semantic Color Output & Spinners**: Modern terminal visual patterns respecting standard `NO_COLOR` directives.
- **Diagnostics Support Engine**: Implemented `orion doctor` to verify configuration directories and troubleshoot offline targets.
- **Comprehensive Unit & Integration Test Suite**: Complete coverage for config serialization, strings sanitization, and CLI command execution flows.
- **CI/CD Workflow**: Automated pull request checks for format, build stability, and test correctness on Ubuntu, macOS, and Windows runners.
- **Documentation Suite**: Added philosophy, architecture guides, troubleshooting protocols, and examples.
