# Testing

Run unit tests and linters before submitting a pull request:
```bash
go vet ./...
go test ./...
golangci-lint run ./...
```

Linux backend tests stub out libsecret so they work on other platforms.
