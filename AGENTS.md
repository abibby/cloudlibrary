# Repository Guidelines

## Project Structure & Module Organization

- `main.go` coordinates authentication, borrowed-book retrieval, and downloads.
- `cloudlibrary/` implements the HTTP client, login, search, book listings, and audiobook/ACSM downloads.
- `libgourou/` wraps external Adobe activation and EPUB utilities.
- Tests live beside implementations in `*_test.go` files.
- `cloudlibrary-audiobook-download.js` is a standalone JavaScript download helper.
- `Dockerfile` builds the Go executable and packages the libgourou runtime; `compose.yaml` configures local execution. `books/` holds downloads, and `data/adept/` holds activation state.

## Build, Test, and Development Commands

Use Go 1.24.4 or newer, matching `go.mod`.

- `go build ./...` — compile all packages without running downloads.
- `go test ./...` — run the complete test suite.
- `go test ./cloudlibrary -run TestLogin` — run focused login tests.
- `go test -cover ./...` — report package coverage.
- `go vet ./...` — run Go's static checks.
- `gofmt -w main.go cloudlibrary/*.go libgourou/*.go` — format Go sources.
- `docker compose run --build --rm cloudlibrary` — build and run with `.env` and mounted data directories.

For native execution, `go run .` requires `adept_activate`, `acsmdownloader`, and `adept_remove` on `PATH`, plus writable `/data` and `/books` paths.

## Coding Style & Naming Conventions

Follow `gofmt` formatting, including tab indentation and grouped imports. Use lowercase package names, PascalCase exported identifiers, and camelCase private identifiers. Follow existing descriptive filenames such as `download-ebook.go`. Wrap errors with operation context and `%w`; keep HTTP behavior in `cloudlibrary/` and subprocess integration in `libgourou/`. No custom lint configuration is present.

## Testing Guidelines

Use the standard `testing` package and `TestBehavior` names. Prefer table-driven subtests for failures. Mock HTTP transports and external commands; use `t.TempDir()` and `t.Setenv()` to isolate state. Command-wrapper tests use POSIX shell scripts and skip relevant cases on Windows. Cover error propagation and preservation of existing downloads. No minimum coverage percentage is configured.

## Commit & Pull Request Guidelines

History currently contains one short, lowercase descriptive commit (`downloading audiobooks and acsm`), with no enforced convention. Write concise, action-oriented subjects and keep commits focused. In PRs, explain behavior changes, link relevant issues, and report validation commands and failures. Include sanitized output when it helps demonstrate CLI changes.

## Security & Configuration

Configure `CLOUDLIBRARY_LIBRARY`, `CLOUDLIBRARY_USERNAME`, and `CLOUDLIBRARY_PASSWORD` through `.env` or environment variables. Never include credentials, session cookies, or downloaded books in commits or logs. Treat `.adept/` and `data/adept/` as sensitive runtime state; avoid adding or modifying activation artifacts in contributions.
