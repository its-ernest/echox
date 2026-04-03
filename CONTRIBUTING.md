# Contributing to echox

First off, thank you for considering contributing to `echox`! It is through developers like you that this tool becomes better for the entire Echo community.

## Versioning Policy

`echox` follows Semantic Versioning (SemVer) and tracks Echo v5 releases closely.

- **Breaking Upstream Changes:** If a new version of Echo v5 breaks the current middleware, we create a legacy branch (e.g., `v5.0-legacy`) and update `main` to support the new API.
- **Bug Fixes:** Should be submitted against the `main` branch.

## <i class="fas fa-tasks"></i> Development Workflow

1. **Fork the Repository:** Create your own fork and clone it locally.
2. **Environment:** Ensure you are using **Go 1.25+**.
3. **Directory Structure:** - New middlewares go in their own top-level directory (e.g., `cache/`).
   - Shared logic (like new store types) goes in `internal/store/`.
4. **Testing:** All new features must include unit tests.
   ```bash
   # use local testing
   gomarkdown -o {{.Dir}}/DOCS.md ./...
   go test -v ./...

   # or use dagger testing
   go run ./dagger/docs/main.go
   go run ./dagger/ci/main.go
   ```


##  Pull Request Guidelines

* **Atomic Commits:** Keep your changes focused. One feature/fix per PR.
* **Documentation:** Update the `README.md` and provide inline GoDoc comments for public functions.
* **Formatting:** Run `go fmt ./...` before committing.
* **Wait for Review:** We aim to review PRs within 48 hours.

##  Reporting Issues

Please use the GitHub Issue Tracker. When reporting a bug, include:

* Your Go version (`go version`)
* Your Echo version
* A minimal reproducible code example (repro)

---

**Built with  for the Go community.**