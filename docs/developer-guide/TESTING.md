# Testing

Run unit tests and linters before submitting a pull request:

```bash
go vet ./...
go test ./...
golangci-lint run ./...
```

Linux backend tests mock libsecret so they can run on any OS.
