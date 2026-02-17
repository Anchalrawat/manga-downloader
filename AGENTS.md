# Agentic Coding Guidelines

This document provides essential information for AI agents and developers working on the `go-manga-ripper` codebase.

## Project Overview

`go-manga-ripper` is a high-performance manga downloader written in Go.
- **Entry Point:** `main.go`
- **Architecture:** Single-file Go application acting as an HTTP server with API endpoints.
- **Key Libraries:** 
  - `github.com/PuerkitoBio/goquery` (HTML parsing)
  - `github.com/valyala/fasthttp` (Fast HTTP client)
- **Functionality:** Fetches manga details, downloads chapters concurrently, and creates CBZ archives.

## Build and Run

### Running the Application
To run the server directly:
```bash
go run main.go
```
The server will start on port `8080`.

### Building the Binary
To compile the application:
```bash
go build -o manga-ripper main.go
```

## Testing and Linting

### Testing
Currently, there are no test files in the repository.
- **Command:** `go test ./...`
- **Guideline:** When adding new functionality, create corresponding `_test.go` files. Use standard Go testing `testing` package.

### Linting
- **Standard:** Use `go vet` for basic static analysis.
```bash
go vet ./...
```
- **Formatting:** Code **must** be formatted using `gofmt`.
```bash
gofmt -w .
```

## Code Style and Conventions

### formatting
- **Indentation:** Use tabs (standard Go style).
- **Line Length:** No hard limit, but keep it readable (approx 80-100 chars).
- **Imports:** Group standard library imports first, then third-party libraries.

### Naming
- **Variables/Functions:** Use `CamelCase`. Exported members must start with an uppercase letter (e.g., `DownloadChapter`). Private members start with lowercase (e.g., `sanitizeFilename`).
- **Files:** snake_case is acceptable for non-Go files, but Go files usually follow package conventions. Since this is a single file app, `main.go` is standard.

### Error Handling
- **Pattern:** Explicit error checking is mandatory.
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```
- **Context:** Wrap errors with context using `fmt.Errorf` and `%w`.
- **Logging:** Use `log` or `fmt.Printf` for operational logs, but avoid excessive noise in tight loops.

### Concurrency
- **Goroutines:** Use `go func()` for parallel tasks.
- **Synchronization:** Use `sync.WaitGroup` to wait for completion.
- **Safety:** Use `sync.Mutex` or `sync.atomic` when accessing shared state (like `progressState`).
- **Throttling:** Respect global semaphores (`imageSemaphore`, `chapterSemaphore`) to avoid rate limiting.

### Networking
- **Clients:** Use the global `httpClient` (standard) or `fasthttpClient` (high performance) as appropriate.
- **Timeouts:** Always configure timeouts on requests to prevent hanging.
- **User-Agent:** Always set a valid User-Agent header to mimic a browser.

## Project Structure

- `main.go`: Contains all application logic, API handlers, and downloaders.
- `go.mod` / `go.sum`: Dependency management.
- `output/`: Directory where downloaded manga and CBZ files are stored.
- `out/`: Directory for the frontend static export (if present).

## API Endpoints

- `GET /api/status`: Returns current download progress `DownloadStatus`.
- `POST /api/fetch`: Fetches manga details for a given URL.
- `POST /api/download`: Starts a background download job.
- `GET /api/files`: Lists available CBZ files in `output/`.
- `GET /api/download-zip`: Downloads a specific CBZ file.

## Specific Implementation Notes

1. **Chunked Downloading:** Large images are downloaded in chunks using `Range` headers to maximize speed.
2. **Sanitization:** All filenames from the web must be sanitized using `sanitizeFilename` before filesystem creation.
3. **CORS:** The server includes a CORS middleware to allow requests from any origin (`*`).
4. **Static Files:** The server serves static files from `output/` for downloads and `out/` for the frontend.

## Git Workflow
- **Commits:** Write clear, concise commit messages.
- **Branching:** Work on feature branches if significant changes are made, though direct commits to main are acceptable for this scale.

