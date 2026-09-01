# Build the native game.
build:
    go build -v ./...

# Run all native unit tests.
test:
    go test -v -count=1 ./internal/game ./internal/physics ./internal/platform ./internal/table ./cmd/...

# Run golangci-lint with the repository configuration.
lint:
    GOOS=js GOARCH=wasm golangci-lint run

# Apply safe linter fixes.
lint-fix:
    GOOS=js GOARCH=wasm golangci-lint run --fix

# Format Go and Markdown sources using the same treefmt pipeline as CI.
fmt:
    treefmt . --allow-missing-formatter --no-cache

# Fail when formatting would change a tracked source file.
fmt-check:
    treefmt --allow-missing-formatter --no-cache --fail-on-change

# Rebuild every original game asset.
assets:
    go run ./cmd/genassets

# Verify committed generated assets are current.
assets-check:
    go test ./cmd/genassets

# Verify all references in an existing browser distribution.
verify-dist:
    go run ./cmd/verifydist

# Build the browser distribution.
web:
    ./scripts/build-web.sh

# Serve the built browser game locally.
run-web: web
    python3 -m http.server --directory dist 8080

# Run the same quality gate used by CI.
check: fmt-check lint test assets-check build web

# Remove reproducible build output.
clean:
    rm -rf dist

# Default target.
default: build

fix:
    just lint-fix
    just fmt
